// controllers/webhook_controller.go
package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/jobs"
	"api-arveshop-go/models"
	"api-arveshop-go/websocket"
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// MidtransNotification struct untuk menampung data dari Midtrans
type MidtransNotification struct {
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	GrossAmount       string `json:"gross_amount"`
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	SignatureKey      string `json:"signature_key"`
	MerchantID        string `json:"merchant_id"`
	
	// Untuk VA
	VaNumbers         []struct {
		Bank     string `json:"bank"`
		VaNumber string `json:"va_number"`
	} `json:"va_numbers,omitempty"`
	
	// Untuk QRIS / E-Wallet
	Actions           []struct {
		Name   string `json:"name"`
		Method string `json:"method"`
		URL    string `json:"url"`
	} `json:"actions,omitempty"`
	
	// Untuk Kartu Kredit
	ApprovalCode      string `json:"approval_code,omitempty"`
	FraudStatus       string `json:"fraud_status,omitempty"`
	Currency          string `json:"currency,omitempty"`
}

func HandleMidtransWebhook(c *gin.Context) {
	// Baca body request
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	
	// Restore body untuk dibaca lagi jika perlu
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	// Log untuk debugging
	log.Printf("Midtrans Webhook received: %s", string(bodyBytes))
	
	// Parse JSON
	var notification MidtransNotification
	if err := json.Unmarshal(bodyBytes, &notification); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	
	// Validasi signature key (keamanan)
	if !validateSignature(notification) {
		log.Printf("Invalid signature key for order: %s", notification.OrderID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}
	
	// Cari transaksi berdasarkan OrderID
	var transaction models.Transaction
	if err := config.DB.Where("order_id = ?", notification.OrderID).First(&transaction).Error; err != nil {
		log.Printf("Transaction not found for order_id: %s", notification.OrderID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	
	// Update transaksi berdasarkan notifikasi
	newStatus, err := updateTransactionFromWebhook(&transaction, notification, bodyBytes)
	if err != nil {
		log.Printf("Error updating transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}
	
	// 🟢 AMBIL DATA TERBARU setelah update
	var updatedTransaction models.Transaction
	config.DB.Where("order_id = ?", notification.OrderID).First(&updatedTransaction)
	
	// 🟢 BROADCAST VIA WEBSOCKET dengan data lengkap
	log.Printf("📢 Broadcasting settlement for order %s via WebSocket", notification.OrderID)
	websocket.BroadcastOrderStatusWithData(notification.OrderID, updatedTransaction)
	
	// Trigger Digiflazz jika settlement
	if notification.TransactionStatus == "settlement" || notification.TransactionStatus == "capture" {
		go triggerDigiflazzProcessing(&updatedTransaction)
	}
	
	// Selalu return 200 OK ke Midtrans
	c.JSON(http.StatusOK, gin.H{
		"status":  newStatus,
		"message": "Notification processed",
	})
}

// Validasi signature key dari Midtrans
func validateSignature(notification MidtransNotification) bool {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		log.Println("WARNING: MIDTRANS_SERVER_KEY not set, skipping signature validation")
		return true
	}
	
	// Format: order_id + status_code + gross_amount + server_key
	signatureString := notification.OrderID + notification.StatusCode + 
		notification.GrossAmount + serverKey
	
	// Generate SHA512 hash
	hash := sha512.New()
	hash.Write([]byte(signatureString))
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	
	// Compare with received signature
	return expectedSignature == notification.SignatureKey
}

// Update transaksi berdasarkan data webhook
func updateTransactionFromWebhook(
	transaction *models.Transaction, 
	notification MidtransNotification, 
	rawBody []byte,
) (string, error) {
	// Mapping status Midtrans ke status aplikasi
	newStatus := mapMidtransStatus(notification.TransactionStatus)
	
	// Parse gross amount dari string ke decimal
	grossAmount, err := parseGrossAmount(notification.GrossAmount)
	if err != nil {
		log.Printf("Error parsing gross amount: %v", err)
	}
	
	// Siapkan data update
	updates := map[string]interface{}{
		"payment_status":   newStatus,
		"status_message":   notification.StatusMessage,
		"updated_at":       time.Now(),
	}
	
	// Update TransactionID jika belum ada
	if transaction.TransactionID == nil || *transaction.TransactionID == "" {
		updates["transaction_id"] = notification.TransactionID
	}
	
	// Update gross amount jika berbeda
	if !grossAmount.Equals(transaction.GrossAmount) && grossAmount.GreaterThan(decimal.Zero) {
		updates["gross_amount"] = grossAmount
	}
	
	// Update PaymentType jika belum ada
	if transaction.PaymentType == nil || *transaction.PaymentType == "" {
		updates["payment_type"] = notification.PaymentType
	}
	
	// Simpan raw Midtrans response
	updates["midtrans_response"] = datatypes.JSON(rawBody)
	
	// Jika settlement (sukses), trigger Digiflazz
	if notification.TransactionStatus == "settlement" || notification.TransactionStatus == "capture" {
		// Trigger proses pengiriman ke Digiflazz
		go triggerDigiflazzProcessing(transaction)
	}
	
	// Jika gagal/expired, update status
	if newStatus == "failed" || newStatus == "expired" {
		digiflazzStatus := "Gagal"
		updates["digiflazz_status"] = &digiflazzStatus
	}
	
	// Update ke database
	if err := config.DB.Model(transaction).Updates(updates).Error; err != nil {
		return "", err
	}
	
	log.Printf("Transaction %s updated: payment_status=%s", 
		notification.OrderID, newStatus)
	
	return newStatus, nil
}

// Mapping status Midtrans ke status aplikasi
func mapMidtransStatus(midtransStatus string) string {
	switch midtransStatus {
	case "capture", "settlement":
		return "settlement"
	case "pending":
		return "pending"
	case "deny", "cancel", "expire", "failure":
		return "failed"
	case "refund":
		return "refunded"
	case "partial_refund":
		return "partial_refund"
	default:
		return "unknown"
	}
}

// Parse gross amount dari format Midtrans (string) ke decimal
func parseGrossAmount(grossAmountStr string) (decimal.Decimal, error) {
	// Midtrans format: "10000.00" atau "10000"
	amount, err := decimal.NewFromString(grossAmountStr)
	if err != nil {
		return decimal.Zero, err
	}
	return amount, nil
}

// Trigger proses pengiriman ke Digiflazz
func triggerDigiflazzProcessing(transaction *models.Transaction) {
    log.Printf("Triggering Digiflazz for order: %s", transaction.OrderID)

    job := jobs.NewDigiflazzTopupJob(
        transaction.ID,
        config.DB,
        config.RDB, // redis client kamu
        jobs.DigiflazzConfig{
            Username: os.Getenv("DIGIFLAZZ_USERNAME"),
            ProdKey:  os.Getenv("DIGIFLAZZ_PROD_KEY"),
            BaseURL:  "https://api.digiflazz.com",
        },
    )

    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
        defer cancel()

        if err := job.Handle(ctx); err != nil {
            log.Printf("Digiflazz job error for order %s: %v", transaction.OrderID, err)
        }
    }()
}

// Endpoint untuk testing webhook
func TestMidtransWebhook(c *gin.Context) {
	var notification MidtransNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Log untuk debugging
	log.Printf("Test webhook received: %+v", notification)
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "settlement",
		"message": "Test webhook received",
		"data":    notification,
	})
}

// Endpoint untuk manual update status (admin)
func ManualUpdateStatus(c *gin.Context) {
	var req struct {
		OrderID         string `json:"order_id" binding:"required"`
		PaymentStatus   string `json:"payment_status" binding:"required"`
		DigiflazzStatus string `json:"digiflazz_status"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Cari transaksi
	var transaction models.Transaction
	if err := config.DB.Where("order_id = ?", req.OrderID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	
	// Update status
	updates := map[string]interface{}{
		"payment_status": req.PaymentStatus,
		"updated_at":     time.Now(),
	}
	
	if req.DigiflazzStatus != "" {
		updates["digiflazz_status"] = req.DigiflazzStatus
	}
	
	if err := config.DB.Model(&transaction).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated",
		"data":    transaction,
	})
}


// Struktur yang sesuai dengan payload Digiflazz
type DigiflazzWebhookPayload struct {
	Data struct {
		RefID          string `json:"ref_id"`  // 🔴 INI AKAN BERISI ORDER_ID KITA
		TrxID          string `json:"trx_id"`
		CustomerNo     string `json:"customer_no"`
		BuyerSkuCode   string `json:"buyer_sku_code"`
		Message        string `json:"message"`
		Status         string `json:"status"`
		RC             string `json:"rc"`
		BuyerLastSaldo int    `json:"buyer_last_saldo"`
		SN             string `json:"sn"`
		Price          int    `json:"price"`
		Tele           string `json:"tele"`
		Wa             string `json:"wa"`
	} `json:"data"`
}

func HandleDigiflazzWebhook(c *gin.Context) {
	// Baca body request
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading Digiflazz webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	
	// Restore body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	// Log untuk debugging
	log.Printf("📥 Digiflazz Webhook received: %s", string(bodyBytes))
	
	// Parse JSON
	var payload DigiflazzWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	
	// 🔴 AMBIL ORDER_ID DARI REF_ID
	data := payload.Data
	orderID := data.RefID
	
	if orderID == "" {
		log.Printf("❌ RefID (OrderID) kosong dalam webhook")
		c.JSON(http.StatusBadRequest, gin.H{"error": "RefID is empty"})
		return
	}
	
	log.Printf("📦 Processing webhook for order: %s, status: %s", orderID, data.Status)
	
	// Cari transaksi berdasarkan OrderID
	var transaction models.Transaction
	if err := config.DB.Where("order_id = ?", orderID).First(&transaction).Error; err != nil {
		log.Printf("❌ Transaction not found for order_id: %s", orderID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	
	// Update status berdasarkan webhook
	statusMessage := data.Message
	updates := map[string]interface{}{
		"digiflazz_status":  data.Status,
		"status_message":    &statusMessage,
		"serial_number":     &data.SN,
		"updated_at":        time.Now(),
	}
	
	// Simpan trx_id untuk referensi (optional)
	if data.TrxID != "" {
		updates["transaction_id"] = &data.TrxID
	}
	
	// Update status pembayaran jika perlu
	if data.Status == "Sukses" {
		updates["payment_status"] = "success"
		log.Printf("✅ Transaksi %s sukses via webhook", orderID)
	} else if data.Status == "Gagal" {
		updates["payment_status"] = "failed"
		log.Printf("❌ Transaksi %s gagal via webhook", orderID)
	}
	
	// Update ke database
	if err := config.DB.Model(&transaction).Updates(updates).Error; err != nil {
		log.Printf("❌ Error updating transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}
	
	// Broadcast via WebSocket
	go websocket.BroadcastOrderStatus(orderID)
	
	// Return 200 OK
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Webhook received",
	})
}