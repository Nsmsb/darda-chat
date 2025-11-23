package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	messageService       service.MessageService
	messageReaderService service.MessageReader
}

func NewMessageHandler(messageService service.MessageService, messageReaderService service.MessageReader) *MessageHandler {
	return &MessageHandler{
		messageService:       messageService,
		messageReaderService: messageReaderService,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins — adjust for production
	},
}

// HandleConnections handles WebSocket connections for real-time chatting.
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
	log.Info("User connected", zap.String("user_id", userId))

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
	go handler.forwardMessages(c, userId, ws)

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

		// Process the message event
		if err := handler.processMessageEvent(c, event, userId); err != nil {
			log.Error("Failed to process message event", zap.String("user_id", userId), zap.Error(err))
			// TODO: send error back to client?
			continue
		}
	}
}

// forwardMessages listens for incoming messages for a user and forwards them to the WebSocket.
func (handler *MessageHandler) forwardMessages(c *gin.Context, userId string, ws *websocket.Conn) {
	log := logger.GetFromContext(c)
	ctx := c.Request.Context()

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
}

// processMessageEvent processes a message event received from the WebSocket.
func (handler *MessageHandler) processMessageEvent(c *gin.Context, event model.Event, userId string) error {
	// Prepare logger from context
	log := logger.GetFromContext(c)

	// Checking event type
	if event.Type != model.EventTypeMessage && event.Type != model.EventTypeMessageEvent {
		log.Error("Unsupported event type", zap.String("user_id", userId), zap.String("event_type", string(event.Type)))
		return fmt.Errorf("unsupported event type: %s", event.Type)
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
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		// Validation of Message
		if msg.Destination == "" || msg.Content == "" {
			log.Error("Invalid message: missing destination or content", zap.String("user_id", userId))
			return fmt.Errorf("invalid message: missing destination or content")
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
			event.Content = json.RawMessage(strMsg)
			strEvent, err := json.Marshal(event)
			if err != nil {
				log.Error("Failed to marshal event", zap.String("user_id", userId), zap.Error(err))
			}
			// Sending message via message service
			err = handler.messageService.SendMessage(c.Request.Context(), msg.Destination, string(strEvent))
			if err != nil {
				log.Error("Failed to send message", zap.String("user_id", userId), zap.Error(err))
				return fmt.Errorf("failed to send message: %w", err)
			}
		} else {
			log.Error("Failed to marshal message", zap.String("user_id", userId), zap.Error(err))
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}
	return nil
}

// GetMessages handles HTTP requests to retrieve messages for a conversation.
func (handler *MessageHandler) GetMessages(c *gin.Context) {
	// Prepare logger from context
	log := logger.GetFromContext(c)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get sender from query, user from path and generate conversation ID
	sender := c.Query("id")
	destination := c.Param("user")
	conversationID := utils.GenerateConvId(sender, destination)

	// Getting cursor parameters from request query
	before := c.Query("before")
	after := c.Query("after")

	// Fetch messages using MessageReaderService
	messages, err := handler.messageReaderService.GetMessages(ctx, conversationID, before, after)
	if err != nil {
		log.Error("Failed to get messages", zap.String("conversation", conversationID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get messages",
		})
		return
	}

	// Determine older and newer cursors
	var olderCursor, newerCursor string
	if len(messages) > 0 {
		last := messages[len(messages)-1]
		first := messages[0]
		newerCursor = fmt.Sprintf("%s_%s", last.Timestamp.Format(time.RFC3339Nano), last.ID)
		olderCursor = fmt.Sprintf("%s_%s", first.Timestamp.Format(time.RFC3339Nano), first.ID)
	}

	// Return messages as JSON response
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"before":   olderCursor,
		"after":    newerCursor,
	})
}
