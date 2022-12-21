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
	"secstorage/internal/server/auth"
	authServices "secstorage/internal/server/auth/services"
	"secstorage/internal/storage"
	authStorage "secstorage/internal/storage/auth"
	"strconv"
)

func main() {
	config := mustReadConfig()
	db := storage.MustInitDB(context.Background(), config.DBURL)

	authStore := authStorage.NewAuthStorage(context.Background(), db)
	authService := authServices.NewAuthService(authStore)
	tokenService := auth.NewTokenService(config.Key)
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
	server.Run(context.Background(), authService, tokenService, creds, listen)
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
