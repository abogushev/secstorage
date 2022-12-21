package server

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	pb "secstorage/internal/api/proto"
	. "secstorage/internal/logger"
	"secstorage/internal/server/auth"
)

func Run(
	ctx context.Context,
	authService auth.Service,
	tokenService *auth.TokenService,
	creds credentials.TransportCredentials,
	listen net.Listener,
) {
	server := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterAuthServer(server, auth.NewServer(authService, tokenService))
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
