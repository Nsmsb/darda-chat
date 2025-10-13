package handler

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type MessageHandler struct {
	connections map[*websocket.Conn]bool
	mutex       sync.Mutex
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		connections: make(map[*websocket.Conn]bool),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins — adjust for production
	},
}

func (handler *MessageHandler) HandleConnections(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("upgrade error:", err)
		return
	}
	// Ensure cleanup when the function exits
	defer func() {
		ws.Close()
		handler.mutex.Lock()
		delete(handler.connections, ws)
		handler.mutex.Unlock()
		fmt.Println("client disconnected and removed")
	}()

	// Add connection
	handler.mutex.Lock()
	handler.connections[ws] = true
	handler.mutex.Unlock()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("read error:", err)
			handler.mutex.Lock()
			delete(handler.connections, ws)
			handler.mutex.Unlock()
			break
		}

		// Broadcast message to all connected clients
		handler.mutex.Lock()
		for client := range handler.connections {
			if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
				fmt.Println("broadcast error:", err)
				client.Close()
				delete(handler.connections, client)
			}
		}
		handler.mutex.Unlock()
	}
}
