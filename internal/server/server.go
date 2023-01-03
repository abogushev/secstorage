package server

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	pb "secstorage/internal/api/proto"
	. "secstorage/internal/logger"
	"secstorage/internal/server/interceptors"
	"secstorage/internal/server/modulservers"
	"secstorage/internal/server/services"
)

func Run(
	ctx context.Context,
	authServer *modulservers.AuthServer,
	resourceServer *modulservers.ResourceServer,
	tokenService *services.TokenService,
	creds credentials.TransportCredentials,
	listen net.Listener,
) {
	server := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(interceptors.TokenInterceptor(tokenService)),
		grpc.StreamInterceptor(interceptors.TokenStreamInterceptor(tokenService)),
	)
	pb.RegisterAuthServer(server, authServer)
	pb.RegisterResourcesServer(server, resourceServer)
	Log.Info("server is up")

	go func() {
		if err := server.Serve(listen); err != nil {
			Log.Fatal("error on up server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	server.GracefulStop()
	Log.Info("server stopped")
}
