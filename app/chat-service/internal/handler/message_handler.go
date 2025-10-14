package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/model"
)

type Client struct {
	List map[*websocket.Conn]bool
}

type MessageHandler struct {
	connections map[string]*Client
	mutex       sync.Mutex
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		connections: make(map[string]*Client),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins — adjust for production
	},
}

func (handler *MessageHandler) HandleConnections(c *gin.Context) {
	// Get id of user and handler error
	userId := c.Query("id")
	if userId == "" {
		fmt.Println("User ID is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "User ID is required",
		})
		return
	}
	fmt.Printf("User %s connected\n", userId)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("upgrade error:", err)
		return
	}

	// Ensure cleanup when the function exits
	defer func() {
		ws.Close()
		handler.mutex.Lock()
		connList, ok := handler.connections[userId]
		if !ok {
			delete(connList.List, ws)
		}
		handler.mutex.Unlock()
		fmt.Println("client disconnected and removed")
	}()

	// Add connection
	handler.mutex.Lock()
	connections, ok := handler.connections[userId]
	if !ok {
		connections = &Client{List: make(map[*websocket.Conn]bool)}
		handler.connections[userId] = connections
	}
	connections.List[ws] = true
	handler.mutex.Unlock()

	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("read error:", err)
			break
		}

		var msg model.Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			fmt.Println("unmarshal error:", err)
			continue
		}
		// Adding current time
		msg.Timestamp = time.Now()

		fmt.Println("new msg", msg)

		// TODO: Validation

		// Broadcast message to all connected clients of destination
		handler.mutex.Lock()
		if destConnections, ok := handler.connections[msg.Destination]; ok {
			// Sending to all destination connections
			for client := range destConnections.List {
				if err := client.WriteJSON(msg); err != nil {
					fmt.Println("broadcast error:", err)
					client.Close()
					delete(destConnections.List, client)
				}
			}
		}
		handler.mutex.Unlock()
	}
}
