package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/model"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/service"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/utils"
	"github.com/nsmsb/darda-chat/app/chat-service/pkg/logger"
	"go.uber.org/zap"
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
	// prepare logger from context
	log := logger.GetFromContext(c)

	// Get id of user and handler error
	userId := c.Query("id")
	if userId == "" {
		log.Error("User ID is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}
	log.Error("User connected", zap.String("user_id", userId))

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Failed to upgrade to WebSocket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to upgrade to WebSocket",
		})
		return
	}

	// Ensure cleanup when the function exits
	defer func() {
		ws.Close()
		log.Info("Client disconnected", zap.String("user_id", userId))
	}()

	// Reading received messages
	go func(ctx context.Context) {
		incomingMessages, err := handler.messageService.SubscribeToMessages(ctx, userId)
		log.Info("Subscribing to messages", zap.String("user_id", userId))
		if err != nil {
			log.Error("Failed to subscribe to messages", zap.Error(err))
			return
		}
		log.Info("Subscribed to messages", zap.String("user_id", userId))

		// Making sure to unsubscribe when done
		defer func() {
			// Unsubscribing from messages
			log.Info("Unsubscribing from messages", zap.String("user_id", userId))
			err := handler.messageService.UnsubscribeFromMessages(userId, incomingMessages)
			if err != nil {
				log.Error("Failed to unsubscribe from messages", zap.Error(err))
			}
		}()

		// Listening for incoming messages and forwarding to WebSocket
		log.Info("Messages listener started", zap.String("user_id", userId))
		for {
			select {
			case msg, ok := <-incomingMessages:
				if !ok {
					log.Info("Incoming messages channel closed", zap.String("user_id", userId))
					return
				}
				// Forwarding message to WebSocket
				ws.WriteMessage(websocket.TextMessage, []byte(msg))

			case <-ctx.Done():
				log.Info("Context done, stopping message listener", zap.String("user_id", userId))
				return
			}
		}
	}(c.Request.Context())

	// Sending Messages
	log.Info("Ready to send messages", zap.String("user_id", userId))
	// Loop to continuously read messages to send from WebSocket
	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			// Check if closed error is because user is disconnected
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Info("Connection closed by client", zap.String("user_id", userId))
			} else {
				log.Error("Error reading from WebSocket", zap.String("user_id", userId), zap.Error(err))
			}
			break
		}

		// Unmarshal incoming event
		var event model.Event
		if err := json.Unmarshal(raw, &event); err != nil {
			log.Error("Failed to unmarshal event", zap.String("user_id", userId), zap.Error(err))
			continue
		}
		// Checking event type
		if event.Type != model.EventTypeMessage && event.Type != model.EventTypeMessageEvent {
			log.Error("Unsupported event type", zap.String("user_id", userId), zap.String("event_type", string(event.Type)))
			continue
		}

		// Creating Id and Timestamp for MessageEvent
		event.Timestamp = time.Now().UTC()
		event.EventID = uuid.New().String()
		// Handling Message event
		if event.Type == model.EventTypeMessage {
			// Unmarshal message content
			var msg model.Message
			if err := json.Unmarshal(event.Content, &msg); err != nil {
				log.Error("Failed to unmarshal message", zap.String("user_id", userId), zap.Error(err))
				continue
			}
			// Validation of Message
			if msg.Destination == "" || msg.Content == "" {
				log.Error("Invalid message: missing destination or content", zap.String("user_id", userId))
				// TODO: Send error back to client?
				continue
			}

			// Adding current time in UTC to avoid server-local timezone differences
			msg.Timestamp = event.Timestamp
			// Setting ID if not provided by the client
			if msg.ID == "" {
				msg.ID = event.EventID
			}
			// Generate conversation ID based on sender and destination
			msg.ConversationID = utils.GenerateConvId(msg.Sender, msg.Destination)

			// Sending Message to Destination
			if strMsg, err := json.Marshal(msg); err == nil {
				// Wrapping message in event
				event := model.Event{
					Type:    model.EventTypeMessage,
					Content: json.RawMessage(strMsg),
				}
				strEvent, err := json.Marshal(event)
				if err != nil {
					log.Error("Failed to marshal event", zap.String("user_id", userId), zap.Error(err))
				}
				// Sending message via message service
				err = handler.messageService.SendMessage(c.Request.Context(), msg.Destination, string(strEvent))
				if err != nil {
					log.Error("Failed to send message", zap.String("user_id", userId), zap.Error(err))
					// TODO: Send error back to client?
					continue
				}
			} else {
				log.Error("Failed to marshal message", zap.String("user_id", userId), zap.Error(err))
				// TODO: Send error back to client?
			}
		}
	}
}
