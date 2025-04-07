.PHONY: build 
build:
	@go build -o ./bin/casper ./cmd/casper

.PHONY: run
run: build
	@./bin/casper