package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

// ChainUnaryInterceptors chains multiple gRPC unary server interceptors into a single interceptor.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// Wrap the handler by iterating backwards
		chainedHandler := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chainedHandler

			chainedHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
				return interceptor(ctx, req, info, next)
			}
		}

		return chainedHandler(ctx, req)
	}
}

// ChainStreamInterceptors chains multiple gRPC stream server interceptors into a single interceptor.
func ChainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		// Wrap the handler by iterating backwards
		chainedHandler := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chainedHandler

			chainedHandler = func(srv interface{}, ss grpc.ServerStream) error {
				return interceptor(srv, ss, info, next)
			}
		}

		return chainedHandler(srv, ss)
	}
}
