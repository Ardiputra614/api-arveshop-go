package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"api-arveshop-go/websocket"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type CreateTransactionRequest struct {
	ID            uint    `json:"id" binding:"required"`
	ProductName   string  `json:"product_name" binding:"required"`
	BuyerSkuCode  string  `json:"buyer_sku_code" binding:"required"`
	CustomerNo    string  `json:"customer_no" binding:"required"`
	SellingPrice  float64 `json:"selling_price" binding:"required"`
	Fee           float64 `json:"fee"`
	PaymentMethodName string  `json:"payment_method_name" binding:"required"`
	WaPembeli     string  `json:"wa_pembeli" binding:"required"`
	ProductType   string  `json:"product_type"`
	PurchasePrice float64 `json:"purchase_price"` // Tambahkan purchase price
}

func CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// ===============================
	// INIT
	// ===============================

	rand.Seed(time.Now().UnixNano())

	// Konversi ke decimal.Decimal
	sellingPrice := decimal.NewFromFloat(req.SellingPrice)
	fee := decimal.NewFromFloat(req.Fee)
	purchasePrice := decimal.NewFromFloat(req.PurchasePrice)
	
	// Hitung gross amount
	grossAmount := sellingPrice.Add(fee)

	orderID := fmt.Sprintf("ORD-%s-%d",
		time.Now().Format("20060102150405"),
		rand.Intn(9000)+1000,
	)

	// ===============================
	// ITEM DETAILS (Midtrans butuh int untuk price)
	// ===============================

	itemDetails := []map[string]interface{}{
		{
			"id":       fmt.Sprintf("%d", req.ID),
			"price":    int(sellingPrice.IntPart()), // Konversi ke int untuk Midtrans
			"quantity": 1,
			"name":     req.ProductName,
		},
	}

	// Tambah fee item jika > 0
	if fee.GreaterThan(decimal.Zero) {
		itemDetails = append(itemDetails, map[string]interface{}{
			"id":       "fee",
			"price":    int(fee.IntPart()),
			"quantity": 1,
			"name":     "Biaya Admin",
		})
	}

	transactionData := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     orderID,
			"gross_amount": int(grossAmount.IntPart()), // Konversi ke int untuk Midtrans
		},
		"item_details": itemDetails,
	}

	paymentType := ""
	paymentMethodName := req.PaymentMethodName

	// ===============================
	// PAYMENT TYPE LOGIC
	// ===============================

	switch req.PaymentMethodName {
	case "qris":
		paymentType = "qris"
		transactionData["payment_type"] = "qris"
		transactionData["qris"] = map[string]interface{}{
			"acquirer": "gopay",
		}

	case "gopay":
		paymentType = "gopay"
		transactionData["payment_type"] = "gopay"
		transactionData["gopay"] = map[string]interface{}{
			"enable_callback": true,
			"callback_url":    os.Getenv("APP_URL") + "/api/callback/midtrans",
		}

	case "shopeepay":
		paymentType = "shopeepay"
		transactionData["payment_type"] = "shopeepay"
		transactionData["shopeepay"] = map[string]interface{}{
			"callback_url": os.Getenv("APP_URL") + "/api/callback/midtrans",
		}

	case "bca", "bni", "bri", "permata", "mandiri", "cimb":
		paymentType = "bank_transfer"
		transactionData["payment_type"] = "bank_transfer"
		transactionData["bank_transfer"] = map[string]interface{}{
			"bank": req.PaymentMethodName,
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment method",
		})
		return
	}

	// ===============================
	// CALL MIDTRANS
	// ===============================

	jsonData, err := json.Marshal(transactionData)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to encode JSON"})
		return
	}

	// Log request untuk debugging
	fmt.Printf("Midtrans Request: %s\n", string(jsonData))

	midtransURL := "https://api.sandbox.midtrans.com/v2/charge"
	if os.Getenv("MIDTRANS_ENV") == "production" {
		midtransURL = "https://api.midtrans.com/v2/charge"
	}

	httpReq, err := http.NewRequest("POST", midtransURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed create request"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.SetBasicAuth(os.Getenv("MIDTRANS_SERVER_KEY"), "")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed call Midtrans: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Log response untuk debugging
	fmt.Printf("Midtrans Response (%d): %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		c.JSON(resp.StatusCode, gin.H{
			"error":   "Midtrans error",
			"message": string(body),
		})
		return
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		c.JSON(500, gin.H{"error": "Invalid Midtrans response"})
		return
	}

	transactionID, _ := responseData["transaction_id"].(string)
	statusMessage, _ := responseData["status_message"].(string)

	urlOrVA := getPaymentURLOrVA(responseData)
	deeplinkGopay := getDeeplinkGopay(responseData)


	// Konversi response ke JSON untuk disimpan di database
	midtransResponseJSON, err := json.Marshal(responseData)
	if err != nil {
		log.Printf("Warning: Failed to marshal Midtrans response: %v", err)
		midtransResponseJSON = []byte("{}") // Default empty object jika error
	}

	// ===============================
	// SAVE TO DATABASE
	// ===============================

	transaction := models.Transaction{
		ProductID:         &req.ID,
		ProductName:       stringPtr(req.ProductName),
		ProductType:       stringPtr(req.ProductType),
		CustomerNo:        req.CustomerNo,
		BuyerSkuCode:      req.BuyerSkuCode,
		OrderID:           orderID,
		TransactionID:     stringPtr(transactionID),
		GrossAmount:       grossAmount,
		SellingPrice:      sellingPrice,
		PurchasePrice:     purchasePrice,
		PaymentType:       stringPtr(paymentType),
		PaymentMethodName: stringPtr(paymentMethodName),
		PaymentStatus:     "pending",
		StatusMessage:     stringPtr(statusMessage),
		URL:               stringPtr(urlOrVA),
		DeeplinkGopay:     stringPtr(deeplinkGopay),
		WaPembeli:         req.WaPembeli,
		MidtransResponse: datatypes.JSON(midtransResponseJSON),
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed save transaction: " + err.Error()})
		return
	}

	// Load relasi jika perlu
	config.DB.Preload("Product").First(&transaction, transaction.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment created",
		"data": gin.H{
			"transaction":   transaction,
			"payment_url":   urlOrVA,
			"deeplink":      deeplinkGopay,
			"midtrans_data": responseData,
		},
	})
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func getPaymentURLOrVA(data map[string]interface{}) string {
	// QRIS / e-wallet
	if actions, ok := data["actions"].([]interface{}); ok {
		for _, a := range actions {
			if action, ok := a.(map[string]interface{}); ok {
				if name, ok := action["name"].(string); ok && name == "generate-qr-code" {
					if url, ok := action["url"].(string); ok {
						return url
					}
				}
				// Fallback ke url pertama
				if url, ok := action["url"].(string); ok {
					return url
				}
			}
		}
	}

	// VA
	if vaNumbers, ok := data["va_numbers"].([]interface{}); ok {
		if len(vaNumbers) > 0 {
			if va, ok := vaNumbers[0].(map[string]interface{}); ok {
				if number, ok := va["va_number"].(string); ok {
					return number
				}
			}
		}
	}

	// Redirect URL
	if redirectURL, ok := data["redirect_url"].(string); ok {
		return redirectURL
	}

	return ""
}

func getDeeplinkGopay(data map[string]interface{}) string {
	if actions, ok := data["actions"].([]interface{}); ok {
		for _, a := range actions {
			if action, ok := a.(map[string]interface{}); ok {
				if name, ok := action["name"].(string); ok && name == "deeplink-redirect" {
					if url, ok := action["url"].(string); ok {
						return url
					}
				}
			}
		}
	}
	return ""
}

func GetStatusPayment(p *gin.Context) {
	orderID := p.Param("order_id")

	var transaction models.Transaction

	err := config.DB.Where("order_id = ?", orderID).First(&transaction).Error

	if err != nil {
		p.JSON(http.StatusNotFound, gin.H{"message": "data tidak ditemukan"})
		return
	}

	p.JSON(http.StatusOK, gin.H{
		"message": "Berhasil",
		"data": gin.H{
			"transaction_id":   transaction.TransactionID,
			"order_id":         transaction.OrderID,
			"payment_status":   transaction.PaymentStatus,
			"digiflazz_status": transaction.DigiflazzStatus,
			"gross_amount":     transaction.GrossAmount,
			"payment_type":     transaction.PaymentType,
			"updated_at":       transaction.UpdatedAt,
		},
	})
}

// WebSocket endpoint
func WebSocketConnection(c *gin.Context) {
	websocket.HandleWebSocket(c)
}

// UpdatePaymentStatus - Contoh fungsi yang memicu WebSocket broadcast
func UpdatePaymentStatus(c *gin.Context) {
	var req struct {
		OrderID        string `json:"order_id"`
		PaymentStatus  string `json:"payment_status"`
		DigiflazzStatus string `json:"digiflazz_status"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Update database
	var transaction models.Transaction
	if err := config.DB.Where("order_id = ?", req.OrderID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	
	updates := map[string]interface{}{
		"payment_status":   req.PaymentStatus,
		"digiflazz_status": req.DigiflazzStatus,
		"updated_at":       time.Now(),
	}
	
	config.DB.Model(&transaction).Updates(updates)
	
	// Broadcast update via WebSocket
	websocket.BroadcastOrderStatus(req.OrderID)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated and broadcasted",
	})
}

// Webhook handler yang memicu WebSocket
// func HandleMidtransWebhook(c *gin.Context) {
// 	// ... existing webhook code ...
	
// 	// After updating transaction status
// 	// websocket.BroadcastOrderStatus(notification.OrderID)
	
// 	c.JSON(http.StatusOK, gin.H{"status": "success"})
// }