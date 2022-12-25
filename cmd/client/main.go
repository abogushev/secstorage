package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "secstorage/internal/api/proto"
	"secstorage/internal/client/services"
	. "secstorage/internal/logger"
	"syscall"
)

var authService *services.AuthService

func main() {
	creds, err := credentials.NewClientTLSFromFile("cert/service.pem", "")
	if err != nil {
		Log.Fatal("could not process the credentials: %v", zap.Error(err))
	}
	con, err := grpc.Dial(":3200", grpc.WithTransportCredentials(creds))
	if err != nil {
		Log.Fatal("create connection failed", zap.Error(err))
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			Log.Error("failed to close connection", zap.Error(err))
		}
	}(con)
	authService = services.NewAuthService(pb.NewAuthClient(con))

	loop()
}

func clear(msg string) {
	fmt.Print("\033[H\033[2J")
	if len(msg) != 0 {
		fmt.Println(msg)
	}
	fmt.Print("-> ")
}

func loop() {
	for {
		cmd := readString("")
		result := processUI(cmd)
		clear(result)
	}
}

func processUI(cmd string) string {
	switch cmd {
	case "login":
		login := readString("input login")
		password := readPassword()
		_, err := authService.Login(context.Background(), login, password)
		if err != nil {
			return err.Error()
		}
		return "login succeed"

	case "register":
		login := readString("input login")
		password := readPassword()
		_, err := authService.Register(context.Background(), login, password)
		if err != nil {
			return err.Error()
		}
		return "register succeed"

	default:
		return `
login - login to account
register - create new account
`
	}
}

func readPassword() string {
	fmt.Println("input password")
	fmt.Print("-> ")
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		panic(err)
	}
	return string(bytePassword)
}

func readString(label string) string {
	if len(label) != 0 {
		fmt.Println(label)
	}
	input := ""
	fmt.Scanln(&input)

	return input
}
