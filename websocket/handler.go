// websocket/handler.go - PERBAIKAN
package websocket

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	
	// Get user ID from query param
	userID := c.Query("user_id")
	
	client := &Client{
		Conn:     conn,
		UserID:   userID,
		OrderIDs: make(map[string]bool),
		Send:     make(chan []byte, 256),
	}
	
	Manager.Register <- client
	
	// Start goroutines
	go client.readPump()
	go client.writePump()
	
	log.Printf("New WebSocket connection established for user: %s", userID)
}

// readPump handles incoming messages from client
func (c *Client) readPump() {
	defer func() {
		Manager.Unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Log received message
		log.Printf("Received message from client %s: %s", c.UserID, string(message))
		
		// Parse client message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}
		
		// Handle different message types
		switch msg.Type {
		case "subscribe":
			c.handleSubscribe(msg)
		case "unsubscribe":
			c.handleUnsubscribe(msg)
		case "get_status":
			c.handleGetStatus(msg)
		case "ping":
			c.handlePing()
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// writePump sends messages to client
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
			
			log.Printf("Sent message to client %s: %s", c.UserID, string(message))
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping: %v", err)
				return
			}
		}
	}
}

// handleSubscribe subscribes client to order updates
func (c *Client) handleSubscribe(msg Message) {
	if msg.OrderID == "" {
		return
	}
	
	c.OrderIDs[msg.OrderID] = true
	log.Printf("Client %s subscribed to order: %s", c.UserID, msg.OrderID)
	
	// Kirim status terkini
	status, err := getOrderStatus(msg.OrderID)
	if err == nil {
		// Gunakan tipe "order_update" untuk konsistensi
		response := Message{
			Type:    "order_update",
			OrderID: msg.OrderID,
			Data:    status,
		}
		jsonResponse, _ := json.Marshal(response)
		c.Send <- jsonResponse
		log.Printf("Sent initial status for order %s to client %s", msg.OrderID, c.UserID)
	} else {
		log.Printf("Error getting order status: %v", err)
	}
}

// handleUnsubscribe unsubscribes client from order updates
func (c *Client) handleUnsubscribe(msg Message) {
	if msg.OrderID == "" {
		return
	}
	
	delete(c.OrderIDs, msg.OrderID)
	log.Printf("Client %s unsubscribed from order: %s", c.UserID, msg.OrderID)
}

// handleGetStatus gets current order status
func (c *Client) handleGetStatus(msg Message) {
	if msg.OrderID == "" {
		return
	}
	
	log.Printf("Client %s requested status for order: %s", c.UserID, msg.OrderID)
	
	status, err := getOrderStatus(msg.OrderID)
	if err != nil {
		response := Message{
			Type:  "error",
			Error: err.Error(),
		}
		jsonResponse, _ := json.Marshal(response)
		c.Send <- jsonResponse
		return
	}
	
	// Gunakan tipe yang SAMA "order_update"
	response := Message{
		Type:    "order_update",
		OrderID: msg.OrderID,
		Data:    status,
	}
	jsonResponse, _ := json.Marshal(response)
	c.Send <- jsonResponse
}

// handlePing responds to ping
func (c *Client) handlePing() {
	response := Message{
		Type: "pong",
	}
	jsonResponse, _ := json.Marshal(response)
	c.Send <- jsonResponse
	log.Printf("Sent pong to client %s", c.UserID)
}

// websocket/handler.go
func getOrderStatus(orderID string) (map[string]interface{}, error) {
	var transaction models.Transaction
	
	err := config.DB.Where("order_id = ?", orderID).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"transaction_id":       transaction.TransactionID,
		"order_id":             transaction.OrderID,
		"payment_status":       transaction.PaymentStatus,
		"digiflazz_status":     transaction.DigiflazzStatus,
		"serial_number":        transaction.SerialNumber,
		"url":                  transaction.URL,
		"gross_amount":         transaction.GrossAmount,
		"payment_type":         transaction.PaymentType,
		"payment_method_name":  transaction.PaymentMethodName,
		"updated_at":           transaction.UpdatedAt,
	}, nil
}

// BroadcastOrderStatus sends status update to all subscribers
func BroadcastOrderStatus(orderID string) {
	status, err := getOrderStatus(orderID)
	if err != nil {
		log.Printf("Error getting order status: %v", err)
		return
	}
	
	log.Printf("Broadcasting status update for order: %s", orderID)
	Manager.SendToOrderSubscribers(orderID, status)
}