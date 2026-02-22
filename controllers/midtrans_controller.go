package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateTransactionRequest struct {
	ID            uint    `json:"id" binding:"required"`
	ProductName   string  `json:"product_name" binding:"required"`
	BuyerSkuCode  string  `json:"buyer_sku_code" binding:"required"`
	CustomerNo    string  `json:"customer_no" binding:"required"`
	SellingPrice  float64 `json:"selling_price" binding:"required"`
	Fee           float64 `json:"fee"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	WaPembeli     string  `json:"wa_pembeli" binding:"required"`
	ProductType   string  `json:"product_type"`
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

	grossAmount := uint64(req.SellingPrice + req.Fee)
	sellingPrice := uint64(req.SellingPrice)

	orderID := fmt.Sprintf("ORD-%s-%d",
		time.Now().Format("20060102150405"),
		rand.Intn(9000)+1000,
	)

	// ===============================
	// ITEM DETAILS
	// ===============================

	itemDetails := []map[string]interface{}{
		{
			"id":       req.ID,
			"price":    int(req.SellingPrice),
			"quantity": 1,
			"name":     req.ProductName,
		},
		{
			"id":       "fee",
			"price":    int(req.Fee),
			"quantity": 1,
			"name":     "Biaya Admin",
		},
	}

	transactionData := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     orderID,
			"gross_amount": grossAmount,
		},
		"item_details": itemDetails,
	}

	paymentType := ""
	paymentMethodName := req.PaymentMethod

	// ===============================
	// PAYMENT TYPE LOGIC
	// ===============================

	switch req.PaymentMethod {

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
			"callback_url":    "https://yourdomain.com/callback",
		}

	case "shopeepay":
		paymentType = "shopeepay"
		transactionData["payment_type"] = "shopeepay"
		transactionData["shopeepay"] = map[string]interface{}{
			"callback_url": "https://yourdomain.com/callback",
		}

	case "bca", "bni", "bri", "permata", "mandiri", "cimb":
		paymentType = "bank_transfer"
		transactionData["payment_type"] = "bank_transfer"
		transactionData["bank_transfer"] = map[string]interface{}{
			"bank": req.PaymentMethod,
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

	midtransURL := "https://api.sandbox.midtrans.com/v2/charge"

	httpReq, err := http.NewRequest("POST", midtransURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed create request"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(os.Getenv("MIDTRANS_SERVER_KEY"), "")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed call Midtrans"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

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
		SellingPrice:      uint64Ptr(sellingPrice),
		PurchasePrice:     uint64Ptr(sellingPrice),
		PaymentType:       stringPtr(paymentType),
		PaymentMethodName: stringPtr(paymentMethodName),
		PaymentStatus:     "pending",
		StatusMessage:     stringPtr(statusMessage),
		URL:               stringPtr(urlOrVA),
		WaPembeli:         req.WaPembeli,
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed save transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment created",
		"data":    responseData,
	})
}


func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}

func getPaymentURLOrVA(data map[string]interface{}) string {

	// QRIS / e-wallet
	if actions, ok := data["actions"].([]interface{}); ok {
		for _, a := range actions {
			if action, ok := a.(map[string]interface{}); ok {
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

	return ""
}


