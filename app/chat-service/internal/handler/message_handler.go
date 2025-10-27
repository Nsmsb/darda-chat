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
		incomingMessages, err := handler.messageService.SubscribeToMessages(ctx, userId)
		fmt.Println("subscribed to messages")
		if err != nil {
			fmt.Println("Error: couldn't subscribe to messages")
			return
		}

		// Making sure to unsubscribe when done
		defer func() {
			fmt.Println("unsubscribing client from messages")
			handler.messageService.UnsubscribeFromMessages(userId, incomingMessages)
		}()

		// Listening for incoming messages and forwarding to WebSocket
		for {
			select {
			case msg, ok := <-incomingMessages:
				if !ok {
					fmt.Println("incoming messages channel closed")
					return
				}
				// Forwarding message to WebSocket
				ws.WriteMessage(websocket.TextMessage, []byte(msg))

			case <-ctx.Done():
				fmt.Println("context done, exiting message listener")
				return
			}
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
		// Adding current time in UTC to avoid server-local timezone differences
		msg.Timestamp = time.Now().UTC()

		fmt.Println("new msg", msg)

		// Sending Message to Destination
		// TODO: Validation
		if strMsg, err := json.Marshal(msg); err == nil {
			// TODO: handle err
			err = handler.messageService.SendMessage(c.Request.Context(), msg.Destination, string(strMsg))
			if err != nil {
				fmt.Println("Error to send message", err)
				continue
			}
		}
	}
}
