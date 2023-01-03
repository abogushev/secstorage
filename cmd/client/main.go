package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
	"secstorage/internal/api"
	pb "secstorage/internal/api/proto"
	"secstorage/internal/client/interceptors"
	"secstorage/internal/client/model"
	"secstorage/internal/client/services"
	. "secstorage/internal/logger"
	"strings"
	"syscall"
)

var authService *services.AuthService
var resourceService *services.ResourceService
var scanner = makeScanner()
var tokenService = &services.TokenService{}

func main() {
	creds, err := credentials.NewClientTLSFromFile("cert/service.pem", "")
	if err != nil {
		Log.Fatal("could not process the credentials: %v", zap.Error(err))
	}
	con, err := grpc.Dial(
		":3200",
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(interceptors.TokenUnaryInterceptor(tokenService)),
		grpc.WithStreamInterceptor(interceptors.TokenStreamInterceptor(tokenService)),
	)
	if err != nil {
		Log.Fatal("create connection failed", zap.Error(err))
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			Log.Error("failed to close connection", zap.Error(err))
		}
	}(con)

	authService = services.NewAuthService(pb.NewAuthClient(con), tokenService)
	resourceService = services.NewResourceService(pb.NewResourcesClient(con))

	loop(loginRegisterInitMsg, initAuth)
	infinityLoop(saveInitMsg, processUI)
}

func clear(msg string) {
	fmt.Print("\033[H\033[2J")
	if len(msg) != 0 {
		fmt.Println(msg)
	}
}

var loginRegisterInitMsg = `
login - to login
register - to register
`

func initAuth(input string) error {
	switch input {
	case "login":
		login := readString("input login")
		password := readPassword()
		_, err := authService.Login(context.Background(), login, password)
		return err

	case "register":
		login := readString("input login")
		password := readPassword()
		_, err := authService.Register(context.Background(), login, password)
		return err
	}
	return errors.New("bad args")
}

func loop(initMsg string, handler func(string) error) {
	clear(initMsg)
	for {
		cmd := readString("")
		err := handler(cmd)
		if err != nil {
			clear(fmt.Sprintf("ERR: %v\n%v", err.Error(), initMsg))
			continue
		}
		return
	}
}

func infinityLoop(initMsg string, handler func(string) (string, error)) {
	clear(initMsg)
	for {
		cmd := readString("")
		if len(cmd) == 0 {
			clear(initMsg)
			continue
		}
		result, err := handler(cmd)
		if err != nil {
			clear(fmt.Sprintf("ERR: %v\n%v", err.Error(), initMsg))
			continue
		}
		clear(fmt.Sprintf("OK: %v\n%v", result, initMsg))
	}
}

func processUI(input string) (string, error) {
	arr := strings.Split(input, " ")

	cmd := arr[0]
	args := arr[1:]

	switch cmd {
	case "save":
		return handleSave(args)
	}
	return "", errors.New("bad args")
}

var saveInitMsg = `
save lp - save login password
save bc - save bank card
`

func handleSave(args []string) (string, error) {
	if args[0] != "lp" && args[0] != "bc" {
		return "", errors.New("bad args")
	}
	var resource any
	var meta string
	var rType api.ResourceType

	switch args[0] {
	case "lp":
		resource, meta = readLoginPassword()
		rType = api.LoginPassword

	case "bc":
		resource, meta = readBankCard()
		rType = api.BankCard
	}

	resourceJson, err := json.Marshal(resource)
	if err != nil {
		return "", err
	}

	id, err := resourceService.Save(context.Background(), rType, resourceJson, []byte(meta))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("saved successfully, id: %v", id), nil
}

func readLoginPassword() (*model.LoginPassword, string) {
	login := readString("input login")
	password := readPassword()
	description := readString("input description")

	return model.NewLoginPassword(login, password), description
}

func readBankCard() (*model.BankCard, string) {
	number := readString("input number")
	until := readString("input until in format: MM/YY")
	name := readString("input name")
	surname := readString("input surname")
	description := readString("input description")

	return model.NewBankCard(number, until, name, surname), description
}

func readPassword() string {
	fmt.Println("input password")
	fmt.Print("-> ")
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		panic(err)
	}
	fmt.Println()
	return string(bytePassword)
}

func makeScanner() *bufio.Scanner {
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(buf, maxCapacity)
	return scanner
}

func readString(label string) string {
	if len(label) != 0 {
		fmt.Println(label)
	}
	fmt.Print("-> ")
	scanner.Scan()
	return scanner.Text()
}
