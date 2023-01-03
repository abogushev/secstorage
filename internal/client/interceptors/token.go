package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TokenService interface {
	Get() string
}

func ctxWithToken(ctx context.Context, tokenService TokenService) context.Context {
	token := tokenService.Get()
	if token != "" {
		return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"token": token}))
	}
	return ctx
}

func TokenUnaryInterceptor(tokenService TokenService) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(ctxWithToken(ctx, tokenService), method, req, reply, cc, opts...)
	}
}

func TokenStreamInterceptor(tokenService TokenService) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(ctxWithToken(ctx, tokenService), desc, cc, method, opts...)
	}
}
