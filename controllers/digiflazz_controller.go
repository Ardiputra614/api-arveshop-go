package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	apiURL := "https://api.digiflazz.com/v1/price-list"
	username := os.Getenv("DIGIFLAZZ_USERNAME")
	apiKey := os.Getenv("DIGIFLAZZ_PROD_KEY")

	sign := md5Hash(username + apiKey + "pricelist")

	payload := map[string]interface{}{
		"cmd":      "prepaid",
		"username": username,
		"sign":     sign,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed create request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed call Digiflazz"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		c.JSON(500, gin.H{
			"error":  "Digiflazz error",
			"body":   string(body),
			"status": resp.StatusCode,
		})
		return
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		c.JSON(500, gin.H{"error": "Invalid JSON response"})
		return
	}

	data, ok := responseData["data"].([]interface{})
	if !ok {
		c.JSON(500, gin.H{
			"error":    "Invalid API response structure",
			"response": responseData,
		})
		return
	}

	// Di dalam fungsi GetProducts Anda
for _, item := range data {
    product, ok := item.(map[string]interface{})
    if !ok {
        continue
    }

    buyerSkuCode := getString(product, "buyer_sku_code")
    productName := getString(product, "product_name")

    if buyerSkuCode == "" || productName == "" {
        continue
    }

    slug := slugify(getString(product, "brand"))

    // 🔴 AMBIL DATA CUTOFF (LANGSUNG STRING)
    startCutOff := getString(product, "start_cut_off")
    endCutOff := getString(product, "end_cut_off")

    var existing models.Product
    err := config.DB.Where("buyer_sku_code = ?", buyerSkuCode).First(&existing).Error

    if err == nil {
        // UPDATE
        updates := map[string]interface{}{
            "product_name":           productName,
            "slug":                   slug,
            "category":               getString(product, "category"),
            "brand":                  getString(product, "brand"),
            "type":                   getString(product, "type"),
            "product_type":           "prepaid",
            "seller_name":            getString(product, "seller_name"),
            "price":                  getUint(product["price"]),
            "selling_price":           getUint(product["price"]),
            "buyer_sku_code":          buyerSkuCode,
            "buyer_product_status":    getBool(product, "buyer_product_status", true),
            "seller_product_status":   getBool(product, "seller_product_status", true),
            "unlimited_stock":         getBool(product, "unlimited_stock", false),
            "multi":                   getBool(product, "multi", false),
            "stock":                   getString(product, "stock"),
            "start_cut_off":           startCutOff, // ✅ LANGSUNG STRING
            "end_cut_off":             endCutOff,   // ✅ LANGSUNG STRING
            "description":             getString(product, "desc"),
            "updated_at":              time.Now(),
        }

        if err := config.DB.Model(&existing).Updates(updates).Error; err != nil {
            log.Printf("Gagal update: %v", err)
        }

    } else {
        // CREATE
        newProduct := models.Product{
            ProductName:         productName,
            Slug:                slug,
            Category:            getString(product, "category"),
            Brand:               getString(product, "brand"),
            Type:                getString(product, "type"),
            ProductType:         "prepaid",
            SellerName:          getString(product, "seller_name"),
            Price:               getUint(product["price"]),
            SellingPrice:        getUint(product["price"]),
            BuyerSkuCode:        buyerSkuCode,
            BuyerProductStatus:  getBool(product, "buyer_product_status", true),
            SellerProductStatus: getBool(product, "seller_product_status", true),
            UnlimitedStock:      getBool(product, "unlimited_stock", false),
            Multi:               getBool(product, "multi", false),
            Stock:               getString(product, "stock"),
            StartCutOff:         startCutOff, // ✅ LANGSUNG STRING
            EndCutOff:           endCutOff,   // ✅ LANGSUNG STRING
            Description:         getString(product, "desc"),
            CreatedAt:           time.Now(),
            UpdatedAt:           time.Now(),
        }

        if err := config.DB.Create(&newProduct).Error; err != nil {
            log.Printf("Gagal create: %v", err)
        }
    }
}

	c.JSON(200, responseData)
}


// func GetProductsPasca(c *gin.Context) {

// 	apiURL := "https://api.digiflazz.com/v1/price-list"
// 	username := os.Getenv("DIGIFLAZZ_USERNAME")
// 	apiKey := os.Getenv("DIGIFLAZZ_PROD_KEY")

// 	sign := md5Hash(username + apiKey + "pricelist")

// 	payload := map[string]interface{}{
// 		"cmd":      "pasca",
// 		"username": username,
// 		"sign":     sign,
// 	}

// 	jsonData, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		c.JSON(500, gin.H{"error": "Failed call Digiflazz"})
// 		return
// 	}
// 	defer resp.Body.Close()

// 	body, _ := io.ReadAll(resp.Body)

// 	var responseData map[string]interface{}
// 	json.Unmarshal(body, &responseData)

// 	data, ok := responseData["data"].([]interface{})
// 	if !ok {
// 		c.JSON(500, gin.H{"error": "Invalid API structure"})
// 		return
// 	}

// 	for _, item := range data {

// 		product, ok := item.(map[string]interface{})
// 		if !ok {
// 			continue
// 		}

// 		buyerSkuCode, _ := product["buyer_sku_code"].(string)
// 		productName, _ := product["product_name"].(string)

// 		if buyerSkuCode == "" || productName == "" {
// 			continue
// 		}

// 		slug := slugify(getString(product, "brand"))

// 		var existing models.ProdukPasca

// 		err := config.DB.Where("buyer_sku_code = ?", buyerSkuCode).First(&existing).Error

// 		if err == nil {
// 			config.DB.Model(&existing).Updates(models.ProdukPasca{
// 				ProductName: productName,
// 				Slug:        slug,
// 				Category:    getString(product, "category"),
// 				Brand:       getString(product, "brand"),
// 				ProductType: "postpaid",
// 				SellerName:  getString(product, "seller_name"),
// 				SellingPrice: getUint(product["price"]),
// 				Price:        getUint(product["price"]),
// 				Admin:        getUint(product["admin"]),
// 				Commission:   getUint(product["commission"]),
// 				UpdatedAt:    time.Now(),
// 			})
// 		} else {
// 			newProduct := models.ProdukPasca{
// 				BuyerSkuCode: buyerSkuCode,
// 				ProductName:  productName,
// 				Slug:         slug,
// 				Category:     getString(product, "category"),
// 				Brand:        getString(product, "brand"),
// 				ProductType:  "postpaid",
// 				SellerName:   getString(product, "seller_name"),
// 				SellingPrice: getUint(product["price"]),
// 				Price:        getUint(product["price"]),
// 				Admin:        getUint(product["admin"]),
// 				Commission:   getUint(product["commission"]),
// 			}
// 			config.DB.Create(&newProduct)
// 		}
// 	}

// 	c.JSON(200, responseData)
// }


func getBool(data map[string]interface{}, key string, defaultValue bool) bool {
    // Cek apakah key ada dan tidak nil
    if val, ok := data[key]; ok && val != nil {
        // Coba konversi ke bool
        if b, ok := val.(bool); ok {
            return b
        }
        
        // Coba konversi dari string
        if str, ok := val.(string); ok {
            strLower := strings.ToLower(str)
            switch strLower {
            case "true", "1", "yes", "aktif", "on":
                return true
            case "false", "0", "no", "tidak", "off":
                return false
            }
        }
        
        // Coba konversi dari number
        if num, ok := val.(float64); ok {
            return num != 0
        }
        if num, ok := val.(int); ok {
            return num != 0
        }
        if num, ok := val.(int64); ok {
            return num != 0
        }
    }
    
    // Jika tidak ada atau tidak valid, kembalikan defaultValue
    return defaultValue
}

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getUint(val interface{}) int64 {
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}



func slugify(text string) string {
	s := strings.ToLower(text)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}