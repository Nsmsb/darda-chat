package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryZapInterceptor returns a gRPC unary server interceptor that logs unary calls.
func UnaryZapInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		start := time.Now()

		// Call the handler
		resp, err = handler(ctx, req)

		// Log details
		st := status.Convert(err)

		logger.Info("grpc unary call",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.String("code", st.Code().String()),
			zap.String("error", st.Message()),
			zap.Any("request", req),
		)

		return resp, err
	}
}

// StreamZapInterceptor returns a gRPC stream server interceptor that logs stream calls.
func StreamZapInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		start := time.Now()

		err := handler(srv, ss)

		st := status.Convert(err)

		logger.Info("grpc stream call",
			zap.String("method", info.FullMethod),
			zap.Bool("is_client_stream", info.IsClientStream),
			zap.Bool("is_server_stream", info.IsServerStream),
			zap.Duration("duration", time.Since(start)),
			zap.String("code", st.Code().String()),
			zap.String("error", st.Message()),
		)

		return err
	}
}
