build_info_flag = -ldflags "-X main.buildVersion=$$(cat cmd/client/version) -X 'main.buildDate=$$(date +'%d/%m/%Y')'"
client_app = cmd/client/main.go

server_up:
	go run cmd/server/main.go -config internal/server/config/local/config.json

env_up:
	docker-compose up -d

env_down:
	docker-compose down

gen_proto:
	rm internal/api/proto/*.go
	protoc --go_out=. --go_opt=paths=source_relative \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/api/proto/*.proto

gen_cert:
	openssl genrsa -out cert/ca.key 4096
	openssl req -new -x509 -key cert/ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 365 -out cert/ca.cert
	openssl genrsa -out cert/service.key 4096
	openssl req -new -key cert/service.key -out cert/service.csr -config cert/cert.conf
	openssl x509 -req -in cert/service.csr -CA cert/ca.cert -CAkey cert/ca.key -CAcreateserial \
		-out cert/service.pem -days 365 -sha256 -extfile cert/cert.conf -extensions req_ext

client_build_windows:
	GOOS=windows GOARCH=amd64 go build -o bin/client/secstorage.exe $(build_info_flag)  $(client_app)

client_build_osx:
	GOOS=darwin GOARCH=arm64 go build -o bin/client/secstorage_osx $(build_info_flag) $(client_app)

client_build_linux:
	GOOS=linux GOARCH=amd64 go build -o bin/client/secstorage_linux $(build_info_flag) $(client_app)