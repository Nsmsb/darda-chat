package server

import (
	pb "github.com/nsmsb/darda-chat/app/message-reader-service/internal/api/message/gen"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/repository"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/server/interceptor"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/service"
	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/logger"
	"google.golang.org/grpc"
)

// NewMessageGRPCServer creates and returns a new gRPC server with registered message service and interceptors.
func NewMessageGRPCServer(conversationRepo repository.ConversationRepository, conversationCacheRepo repository.ConversationCacheRepository) *grpc.Server {
	logger := logger.Get()
	defer logger.Sync()

	// Create server and add interceptors
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.ChainUnaryInterceptors(
				interceptor.UnaryRecoveryInterceptor(logger),
				interceptor.UnaryZapInterceptor(logger),
			),
		),
		grpc.StreamInterceptor(
			interceptor.ChainStreamInterceptors(
				interceptor.StreamRecoveryInterceptor(logger),
				interceptor.StreamZapInterceptor(logger),
			),
		),
	)

	// Creating message service
	messageService := service.NewMessageService(conversationRepo, conversationCacheRepo)

	// Register Message service
	pb.RegisterMessageServiceServer(server, messageService)

	return server
}
