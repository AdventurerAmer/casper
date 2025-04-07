.PHONY: build 
build:
	@go build -o ./bin/casper ./cmd/casper

.PHONY: run
run: build
	@./bin/casper

.PHONY: generate_tls_cert
generate_tls_cert:
	@openssl genrsa -out tls/key.pem 2048
	@openssl req -new -key tls/key.pem -out tls/cert.pem
	@openssl x509 -req -days 365 -in tls/cert.pem -signkey tls/key.pem -out tls/cert.pem
	@openssl x509 -in tls/cert.pem -text -noout