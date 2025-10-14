package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/model"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/service"
)

type Client struct {
	List map[*websocket.Conn]bool
}

type MessageHandler struct {
	messageService service.MessageService
}

func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
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
			"error": "User ID is required",
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
		fmt.Println("client disconnected and removed")
	}()

	// Reading received messages
	go func(ctx context.Context) {
		receivedMessages, err := handler.messageService.SubscribeToMessages(ctx, userId)
		if err != nil {
			fmt.Println("Error: couldn't subscribe to messages")
			return
		}
		fmt.Println("subscribed to messages")
		for receivedMsg := range receivedMessages {
			fmt.Println("received new msg:", receivedMsg)
			ws.WriteMessage(websocket.TextMessage, []byte(receivedMsg))
		}
	}(c.Request.Context())

	// Reading Messages to send
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

		// Sending Message to Destination
		// TODO: Validation
		if strMsg, err := json.Marshal(msg); err == nil {
			// TODO: handle err
			err = handler.messageService.SendMessage(c, msg.Destination, string(strMsg))
			if err != nil {
				fmt.Println("Error to send message")
				continue
			}
		}
	}
}
