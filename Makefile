
server_up:
	go run cmd/server/main.go -config internal/config/local/config.json

env_up:
	docker-compose up

proto_gen:
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
