package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/repository"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type MessageDispatcherService struct {
	repository repository.OutboxMessageRepository
	dispatcher Dispatcher
	ticker     *time.Ticker
}

func NewMessageDispatcherService(dispatcher Dispatcher, repository repository.OutboxMessageRepository) *MessageDispatcherService {
	return &MessageDispatcherService{
		dispatcher: dispatcher,
		repository: repository,
		ticker:     time.NewTicker(1 * time.Second),
	}
}

func (s MessageDispatcherService) Start(ctx context.Context) error {
	log := logger.FromContext(ctx)

	// Declare the exchange
	err := s.dispatcher.DeclareExchange()
	if err != nil {
		log.Error("Failed to declare exchange", zap.Error(err))
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-s.ticker.C:
			err := s.processMessages(ctx)
			if err != nil {
				// TODO: Log the error and send to Dead Letter Queue.
				log.Error("Error processing messages", zap.Error(err))
			}
			log.Info("Processed outbox messages cycle")
		}
	}
}

func (s MessageDispatcherService) processMessages(ctx context.Context) error {
	log := logger.FromContext(ctx)

	messages, err := s.repository.StreamUnprocessedMessages(ctx)
	if err != nil {
		return err
	}

	for msg := range messages {
		// TODO: Add retry logic and error handling

		// Transaction function
		// TODO: Consider batching messages for better performance and scalability
		callback := func(sessionCtx mongo.SessionContext) (interface{}, error) {
			// Mark message as processed
			err = s.repository.MarkMessageAsProcessed(sessionCtx, msg)
			if err != nil {
				return nil, fmt.Errorf("mark message as processed error: %w", err)
			}
			// Dispatch message
			err := s.dispatcher.Dispatch(msg.Payload)
			if err != nil {
				return nil, fmt.Errorf("dispatch message error: %w", err)
			}
			return nil, nil
		}

		// Creating a session for transaction
		session, err := s.repository.Client().StartSession()
		if err != nil {
			return fmt.Errorf("start session error: %w", err)
		}
		defer session.EndSession(ctx)

		// Executing the transaction
		_, err = session.WithTransaction(ctx, callback)
		if err != nil {
			return fmt.Errorf("transaction error: %w", err)
		}
		log.Info("Dispatched outbox message", zap.String("messageID", msg.ID))
	}
	return nil
}
