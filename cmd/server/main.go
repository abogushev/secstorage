package main

import (
	"context"
	"encoding/json"
	"flag"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"os"
	. "secstorage/internal/logger"
	"secstorage/internal/server"
	"secstorage/internal/server/implementations"
	"secstorage/internal/server/services"
	"secstorage/internal/server/storage"
	authStorage "secstorage/internal/server/storage/auth"
	resourceStorage "secstorage/internal/server/storage/resource"

	"strconv"
)

func main() {
	config := mustReadConfig()
	db := storage.MustInitDB(context.Background(), config.DBURL)

	var creds credentials.TransportCredentials
	if config.UseSecCreds {
		secCreds, err := credentials.NewServerTLSFromFile("cert/service.pem", "cert/service.key")
		if err != nil {
			Log.Fatal("Failed to setup TLS: %v", zap.Error(err))
		}
		creds = secCreds
	} else {
		creds = insecure.NewCredentials()
	}

	listen, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		Log.Error("error on listen port", zap.Error(err))
		return
	}

	tokenService := services.NewTokenService(config.Key)

	authStore := authStorage.NewStorage(context.Background(), db)
	authService := services.NewAuthService(authStore)
	authServer := implementations.NewAuthServer(authService, tokenService)

	resourceStore := resourceStorage.NewStore(context.Background(), db)
	resourceService := services.NewResourceStoreService(resourceStore)
	resourceServer := implementations.NewResourcesServer(resourceService)

	server.Run(context.Background(), authServer, resourceServer, tokenService, creds, listen)
}

type Config struct {
	DBURL       string `json:"dburl"`
	Key         string `json:"key"`
	UseSecCreds bool   `json:"use_sec_creds"`
	Port        int    `json:"port"`
}

func mustReadConfig() Config {
	confPath := flag.String("config", "", "path to conf file")
	flag.Parse()
	data, err := os.ReadFile(*confPath)
	if err != nil {
		panic(err)
	}

	var conf Config
	err = json.Unmarshal(data, &conf)

	if err != nil {
		panic(err)
	}

	return conf
}
