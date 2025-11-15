package interceptor

import (
	"context"
	"runtime/debug"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryRecoveryInterceptor returns a gRPC unary server interceptor that recovers from panics.
func UnaryRecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered in gRPC handler",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.ByteString("stack", debug.Stack()),
				)

				// Convert panic to gRPC internal error
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor returns a gRPC stream server interceptor that recovers from panics.
func StreamRecoveryInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {

		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered in stream handler",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.ByteString("stack", debug.Stack()),
				)

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}
