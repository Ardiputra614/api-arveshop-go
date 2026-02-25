// jobs/digiflazz_topup.go
package jobs

import (
	"api-arveshop-go/models"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TaskDigiflazzTopup = "digiflazz:topup"

type DigiflazzTopupPayload struct {
	OrderID uint `json:"order_id"`
}

// ─── Config ───────────────────────────────────────────────────────────────────

type DigiflazzConfig struct {
	Username string
	ProdKey  string
	BaseURL  string
}

// ─── Job ──────────────────────────────────────────────────────────────────────

type DigiflazzTopupJob struct {
	OrderID uint
	db      *gorm.DB
	rdb     *redis.Client
	cfg     DigiflazzConfig

	maxRetries int
	backoff    []time.Duration
}

func NewDigiflazzTopupJob(orderID uint, db *gorm.DB, rdb *redis.Client, cfg DigiflazzConfig) *DigiflazzTopupJob {
	return &DigiflazzTopupJob{
		OrderID:    orderID,
		db:         db,
		rdb:        rdb,
		cfg:        cfg,
		maxRetries: 5,
		backoff:    []time.Duration{60, 180, 300, 600, 900},
	}
}

// Handle adalah entry point job, dipanggil oleh worker
func (j *DigiflazzTopupJob) Handle(ctx context.Context) error {
	// Load & lock order
	var order models.Transaction
	if err := j.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&order, j.OrderID).Error; err != nil {
		slog.Error("Order not found", "order_id", j.OrderID, "err", err)
		return nil // jangan retry kalau order tidak ada
	}

	// Cek apakah sudah sukses
	if order.DigiflazzStatus != nil && (*order.DigiflazzStatus == "Sukses" || *order.DigiflazzStatus == "cancelled") {
		return nil
	}

	// Distributed lock via Redis
	lockKey := fmt.Sprintf("digiflazz_topup_%s", order.OrderID)
	lock, err := j.acquireLock(ctx, lockKey, 300*time.Second)
	if err != nil {
		slog.Warn("Lock tidak bisa didapat", "order_id", order.OrderID)
		return nil
	}
	defer lock.Release(ctx)

	if err := j.processTopup(ctx, &order); err != nil {
		return j.handleException(ctx, &order, err)
	}

	return nil
}

func (j *DigiflazzTopupJob) processTopup(ctx context.Context, order *models.Transaction) error {
	// Ambil data produk untuk cek cutoff
	var product models.Product
	productErr := j.db.First(&product, order.ProductID).Error
	if productErr == nil {
		// Cek cutoff
		if product.IsWithinCutoff() {
			statusMsg := "Produk sedang cutoff"
			retryAt := time.Now().Add(10 * time.Minute)
			return j.db.Model(order).Updates(map[string]any{
				"digiflazz_status": "pending",
				"status_message":   &statusMsg,
				"retry_at":         &retryAt,
			}).Error
		}
	} else {
		slog.Warn("Product not found", "product_id", order.ProductID, "err", productErr)
	}

	// Debit saldo sekali saja
	if order.SaldoDebitedAt == nil {
		if err := j.debitSaldo(ctx, order); err != nil {
			return err
		}
		// Reload setelah update
		if err := j.db.First(order, order.ID).Error; err != nil {
			return err
		}
		// Kalau debit gagal, order sudah diset failed
		if order.DigiflazzStatus != nil && *order.DigiflazzStatus == "failed" {
			return nil
		}
	}

	return j.hitDigiflazzAPI(ctx, order)
}

// ─── Saldo ────────────────────────────────────────────────────────────────────

func (j *DigiflazzTopupJob) debitSaldo(ctx context.Context, order *models.Transaction) error {
	return j.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var profil models.ProfilAplikasi
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&profil).Error; err != nil {
			slog.Error("ProfilAplikasi tidak ditemukan")
			statusMsg := "Konfigurasi aplikasi tidak ditemukan"
			lastErr := "NOPROF"
			return tx.Model(order).Updates(map[string]any{
				"digiflazz_status": "failed",
				"status_message":   &statusMsg,
				"last_error_code":  &lastErr,
			}).Error
		}

		// Konversi PurchasePrice (decimal.Decimal) ke float64
		purchasePrice, _ := order.PurchasePrice.Float64()

		if profil.Saldo < purchasePrice {
			slog.Error("Saldo aplikasi tidak mencukupi",
				"saldo_tersedia", profil.Saldo,
				"saldo_dibutuhkan", purchasePrice,
			)
			statusMsg := "Saldo aplikasi tidak mencukupi"
			lastErr := "INSUFF"
			return tx.Model(order).Updates(map[string]any{
				"digiflazz_status": "failed",
				"status_message":   &statusMsg,
				"last_error_code":  &lastErr,
			}).Error
		}

		saldoSebelum := profil.Saldo
		if err := tx.Model(&profil).UpdateColumn("saldo", gorm.Expr("saldo - ?", purchasePrice)).Error; err != nil {
			return err
		}

		slog.Info("Saldo dipotong",
			"order_id", order.OrderID,
			"saldo_sebelum", saldoSebelum,
			"dipotong", purchasePrice,
		)

		now := time.Now()
		statusMsg := "Saldo dipotong, memproses transaksi..."
		return tx.Model(order).Updates(map[string]any{
			"saldo_debited_at": &now,
			"digiflazz_status": "processing",
			"status_message":   &statusMsg,
		}).Error
	})
}

func (j *DigiflazzTopupJob) refundSaldo(ctx context.Context, order *models.Transaction) error {
	return j.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var profil models.ProfilAplikasi
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&profil).Error; err != nil {
			slog.Error("ProfilAplikasi tidak ditemukan untuk refund")
			return nil
		}

		purchasePrice, _ := order.PurchasePrice.Float64()
		saldoSebelum := profil.Saldo
		
		if err := tx.Model(&profil).UpdateColumn("saldo", gorm.Expr("saldo + ?", purchasePrice)).Error; err != nil {
			return err
		}

		slog.Info("Saldo dikembalikan",
			"order_id", order.OrderID,
			"saldo_sebelum", saldoSebelum,
			"dikembalikan", purchasePrice,
		)
		return nil
	})
}

// ─── API ──────────────────────────────────────────────────────────────────────

type digiflazzPayload struct {
	Username     string `json:"username"`
	BuyerSkuCode string `json:"buyer_sku_code"`
	CustomerNo   string `json:"customer_no"`
	RefID        string `json:"ref_id"`
	Sign         string `json:"sign"`
}

type digiflazzResponseData struct {
	RC      string `json:"rc"`
	Message string `json:"message"`
	SN      string `json:"sn"`
	RefID   string `json:"ref_id"`
}

type digiflazzResponse struct {
	Data digiflazzResponseData `json:"data"`
}

func (j *DigiflazzTopupJob) hitDigiflazzAPI(ctx context.Context, order *models.Transaction) error {
	payload := j.buildPayload(order)

	slog.Info("Mengirim request ke Digiflazz", "order_id", order.OrderID)

	now := time.Now()
	j.db.Model(order).Update("digiflazz_sent_at", &now)

	payloadJSON, _ := json.Marshal(payload)

	// Tentukan timeout berdasarkan produk
	var timeout time.Duration = 30 * time.Second
	var product models.Product
	if err := j.db.First(&product, order.ProductID).Error; err == nil {
		timeout = j.getAPITimeout(&product)
	}

	httpCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := j.doHTTPPost(httpCtx, j.cfg.BaseURL+"/v1/transaction", payloadJSON)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}

	// Simpan request & response ke DB
	j.db.Model(order).Updates(map[string]any{
		"digiflazz_request":  payloadJSON,
		"digiflazz_response": resp,
	})

	var apiResp digiflazzResponse
	if err := json.Unmarshal(resp, &apiResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return j.handleAPIResponse(ctx, order, apiResp.Data)
}

func (j *DigiflazzTopupJob) handleAPIResponse(ctx context.Context, order *models.Transaction, data digiflazzResponseData) error {
	rc := data.RC
	message := data.Message
	if message == "" {
		message = "Unknown response"
	}

	slog.Info("Response dari Digiflazz", "order_id", order.OrderID, "rc", rc, "message", message)

	switch rc {
	case "00":
		return j.handleSuccess(order, data)
	case "201":
		return j.handlePending(order, message)
	case "40", "41", "42", "43", "44", "45":
		return j.handleFailed(ctx, order, message, rc)
	case "06", "07", "08", "17", "39":
		return j.handleRetryable(ctx, order, message, rc)
	default:
		return j.handleUnknown(order, message, rc)
	}
}

func (j *DigiflazzTopupJob) handleSuccess(order *models.Transaction, data digiflazzResponseData) error {
	status := "Sukses"
	err := j.db.Model(order).Updates(map[string]any{
		"digiflazz_status": &status,
		"status_message":   "Transaksi berhasil",
		"serial_number":    &data.SN,
		"ref_id":           &data.RefID,
	}).Error
	if err == nil {
		slog.Info("✅ Transaksi sukses", "order_id", order.OrderID, "sn", data.SN)
	}
	return err
}

func (j *DigiflazzTopupJob) handlePending(order *models.Transaction, message string) error {
	status := "pending"
	err := j.db.Model(order).Updates(map[string]any{
		"digiflazz_status": &status,
		"status_message":   &message,
	}).Error
	if err == nil {
		slog.Info("⏳ Menunggu callback", "order_id", order.OrderID)
	}
	return err
}

func (j *DigiflazzTopupJob) handleFailed(ctx context.Context, order *models.Transaction, message, rc string) error {
	// Refund saldo jika sudah didebit
	if order.SaldoDebitedAt != nil {
		if err := j.refundSaldo(ctx, order); err != nil {
			slog.Error("Gagal refund saldo", "order_id", order.OrderID, "err", err)
		} else {
			purchasePrice, _ := order.PurchasePrice.Float64()
			slog.Info("💸 Saldo dikembalikan", "order_id", order.OrderID, "amount", purchasePrice)
		}
	}

	status := "failed"
	err := j.db.Model(order).Updates(map[string]any{
		"digiflazz_status": &status,
		"status_message":   &message,
		"last_error_code":  &rc,
	}).Error
	if err == nil {
		slog.Error("❌ Transaksi gagal", "order_id", order.OrderID, "rc", rc)
	}
	return err
}

func (j *DigiflazzTopupJob) handleRetryable(ctx context.Context, order *models.Transaction, message, rc string) error {
	// Increment retry count
	j.db.Model(order).UpdateColumn("retry_count", gorm.Expr("retry_count + 1"))
	j.db.Select("retry_count").First(order, order.ID)

	if order.RetryCount >= j.maxRetries {
		return j.handleFailed(ctx, order, "Gagal setelah 5x retry", rc)
	}

	status := "pending"
	retryAt := time.Now().Add(10 * time.Minute)
	err := j.db.Model(order).Updates(map[string]any{
		"digiflazz_status": &status,
		"status_message":   &message,
		"last_error_code":  &rc,
		"retry_at":         &retryAt,
	}).Error

	slog.Warn("⚠️ Retry transaksi", "order_id", order.OrderID, "retry_count", order.RetryCount, "rc", rc)
	return err
}

func (j *DigiflazzTopupJob) handleUnknown(order *models.Transaction, message, rc string) error {
	status := "pending"
	retryAt := time.Now().Add(10 * time.Minute)
	err := j.db.Model(order).Updates(map[string]any{
		"digiflazz_status": &status,
		"status_message":   &message,
		"last_error_code":  &rc,
		"retry_at":         &retryAt,
	}).Error

	slog.Error("❓ Response code tidak dikenali", "order_id", order.OrderID, "rc", rc)
	return err
}

func (j *DigiflazzTopupJob) handleException(ctx context.Context, order *models.Transaction, e error) error {
	slog.Error("Exception saat processing", "order_id", order.OrderID, "err", e)

	// Update sent_at dan retry_count
	now := time.Now()
	j.db.Model(&models.Transaction{}).Where("id = ?", order.ID).Update("digiflazz_sent_at", &now)
	j.db.Model(&models.Transaction{}).Where("id = ?", order.ID).UpdateColumn("retry_count", gorm.Expr("retry_count + 1"))

	// Reload retry_count
	j.db.Select("retry_count").First(order, order.ID)

	if order.RetryCount >= j.maxRetries {
		return j.handleFailed(ctx, order, "Error: "+e.Error(), "EXCEPT")
	}

	status := "pending"
	statusMsg := "Gangguan sistem"
	retryAt := time.Now().Add(10 * time.Minute)
	j.db.Model(order).Updates(map[string]any{
		"digiflazz_status": &status,
		"status_message":   &statusMsg,
		"retry_at":         &retryAt,
	})

	return e // propagate ke worker supaya bisa reschedule
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (j *DigiflazzTopupJob) buildPayload(order *models.Transaction) digiflazzPayload {
	sign := fmt.Sprintf("%x", md5.Sum([]byte(j.cfg.Username+j.cfg.ProdKey+order.OrderID)))
	return digiflazzPayload{
		Username:     j.cfg.Username,
		BuyerSkuCode: order.BuyerSkuCode,
		CustomerNo:   order.CustomerNo,
		RefID:        order.OrderID,
		Sign:         sign,
	}
}

func (j *DigiflazzTopupJob) getAPITimeout(product *models.Product) time.Duration {
	if product == nil {
		return 30 * time.Second
	}
	
	slowProviders := map[string]bool{
		"PLN":   true,
		"BPJS":  true,
		"TELKOM": true,
		"PASCABAYAR": true,
	}
	
	if slowProviders[product.Category] {
		return 60 * time.Second
	}
	return 30 * time.Second
}


// doHTTPPost melakukan HTTP POST request
func (j *DigiflazzTopupJob) doHTTPPost(ctx context.Context, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return io.ReadAll(resp.Body)
}

// jobs/digiflazz_topup.go - Bagian Lock

// SimpleLock adalah implementasi lock sederhana
type SimpleLock struct {
	key     string
	rdb     *redis.Client
	ctx     context.Context
}

// acquireLock mendapatkan distributed lock via Redis
func (j *DigiflazzTopupJob) acquireLock(ctx context.Context, key string, ttl time.Duration) (*SimpleLock, error) {
	// Gunakan SETNX untuk mendapatkan lock
	success, err := j.rdb.SetNX(ctx, key, "locked", ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}
	
	if !success {
		return nil, fmt.Errorf("lock already acquired for key: %s", key)
	}
	
	// Lock berhasil didapat
	return &SimpleLock{
		key: key,
		rdb: j.rdb,
		ctx: ctx,
	}, nil
}

// Release melepas lock dengan menghapus key dari Redis
func (l *SimpleLock) Release(ctx context.Context) error {
	if l == nil || l.rdb == nil {
		return nil
	}
	
	// Hapus key lock
	_, err := l.rdb.Del(ctx, l.key).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	
	return nil
}