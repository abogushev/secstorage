package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"secstorage/internal/server/services"
)

type ServerStreamWithCtx struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *ServerStreamWithCtx) Context() context.Context {
	return w.ctx
}

func isAuthMethod(method string) bool {
	return method == "/secstorage.Auth/Register" || method == "/secstorage.Auth/Login"
}

func TokenInterceptor(tokenService *services.TokenService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if !isAuthMethod(info.FullMethod) {
			userId, err := tokenService.GetUserIdGRPC(ctx)
			if err != nil {
				return nil, err
			}
			ctxWithUserId := context.WithValue(ctx, "userId", userId)
			return handler(ctxWithUserId, req)
		}
		return handler(ctx, req)
	}
}

func TokenStreamInterceptor(tokenService *services.TokenService) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !isAuthMethod(info.FullMethod) {
			userId, err := tokenService.GetUserIdGRPC(ss.Context())
			if err != nil {
				return err
			}
			ctx := context.WithValue(ss.Context(), "userId", userId)
			ssCtx := &ServerStreamWithCtx{ServerStream: ss, ctx: ctx}

			return handler(srv, ssCtx)
		}
		return handler(srv, ss)
	}
}
