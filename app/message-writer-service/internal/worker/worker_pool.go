package worker

import (
	"context"
	"sync"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/processor"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/source"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	"go.uber.org/zap"
)

type WorkerPool[T any] struct {
	source    source.Source[T]
	processor processor.Processor[T]
	poolSize  int
	workers   chan struct{}
	wg        sync.WaitGroup
}

func NewWorkerPool[T any](source source.Source[T], processor processor.Processor[T], size int) *WorkerPool[T] {
	return &WorkerPool[T]{
		poolSize:  size,
		source:    source,
		processor: processor,
		workers:   make(chan struct{}, size),
	}
}

func (wp *WorkerPool[T]) Start(ctx context.Context) error {
	log := logger.FromContext(ctx)

	// Declare the source (e.g., RabbitMQ queue)
	log.Info("Declaring source queue")
	err := wp.source.DeclareQueue(ctx)
	if err != nil {
		log.Error("Failed to declare source", zap.Error(err))
		return err
	}
	log.Info("Source queue declared successfully")

	events, err := wp.source.Events(ctx)
	if err != nil {
		log.Error("Failed to get events", zap.Error(err))
		return err
	}

	log.Info("Starting worker pool", zap.Int("poolSize", wp.poolSize))
	for {
		select {
		case <-ctx.Done():

			return ctx.Err()
		case EventEnvelope := <-events:
			// Acquire a worker slot.
			wp.workers <- struct{}{}
			wp.wg.Add(1)

			// A worker slot is available, start a new worker.
			go func(eventEnvelope source.EventEnvelope[T]) {
				// Ensure the worker slot is released when done.
				defer func() {
					if r := recover(); r != nil {
						log.Error("Recovered in message processing", zap.Any("error", r))
						_ = wp.source.Nack(eventEnvelope.DeliveryTag, true)
					}
					// Release the worker slot by reading from the channel
					<-wp.workers
					wp.wg.Done()
				}()

				// Process the event & log any errors.
				err := wp.processor.Process(ctx, eventEnvelope.Payload)
				if err != nil {
					log.Error("Error processing event", zap.String("event_id", eventEnvelope.ID), zap.Error(err))
					wp.source.Nack(EventEnvelope.DeliveryTag, true)
					return
				}
				// Acknowledge the event upon successful processing.
				err = wp.source.Ack(EventEnvelope.DeliveryTag)
				if err != nil {
					log.Error("Error acknowledging event", zap.String("event_id", eventEnvelope.ID), zap.Error(err))
					return
				}
				log.Info("Successfully processed event", zap.String("event_id", eventEnvelope.ID))
			}(EventEnvelope)
		}
	}
}

func (wp *WorkerPool[T]) Stop() error {
	log := logger.Get()
	log.Info("Stopping worker pool")
	wp.wg.Wait()
	log.Info("Worker pool stopped")
	return nil
}
