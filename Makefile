.PHONY: build run test tidy

build:
	go build -v ./...

run:
	go run ./

test:
	go test ./... -v

tidy:
	go mod tidy
