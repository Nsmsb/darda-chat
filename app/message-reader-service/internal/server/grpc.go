package server

import (
	pb "github.com/nsmsb/darda-chat/app/message-reader-service/internal/api/message/gen"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/repository"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/server/interceptor"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/service"
	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

// NewMessageGRPCServer creates and returns a new gRPC server with registered message service and interceptors.
func NewMessageGRPCServer(mongoClient *mongo.Client, redisClient *redis.Client) *grpc.Server {
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

	// Preparing for message service creation
	conversationRepo := repository.NewMongoConversationRepository(mongoClient)
	conversationCacheRepo := repository.NewRedisConversationCacheRepository(redisClient)

	// Creating message service
	messageService := service.NewMessageService(conversationRepo, conversationCacheRepo)

	// Register Message service
	pb.RegisterMessageServiceServer(server, messageService)

	return server
}
