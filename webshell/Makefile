# Makefile for webshell server

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=server
BINARY_UNIX=$(BINARY_NAME)_unix

# Server parameters
HOST ?= 0.0.0.0
PORT ?= 8080

# Build targets
.PHONY: all build clean test deps swagger run dev install docker-build help

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v server.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v server.go

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME).exe -v server.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f $(BINARY_UNIX)

test:
	$(GOTEST) -v ./...

deps:
	$(GOMOD) download
	$(GOMOD) tidy

swagger:
	swag init -g server.go

run: build
	./$(BINARY_NAME) -host $(HOST) -port $(PORT)

run-windows: build-windows
	$(BINARY_NAME).exe -host $(HOST) -port $(PORT)

dev:
	$(GOCMD) run server.go -host $(HOST) -port $(PORT)

# 特定端口运行
run-8080:
	$(MAKE) run PORT=8080

run-9090:
	$(MAKE) run PORT=9090

run-local:
	$(MAKE) run HOST=127.0.0.1 PORT=8080

run-prod:
	$(MAKE) run HOST=0.0.0.0 PORT=80

install:
	$(GOGET) github.com/swaggo/swag/cmd/swag
	$(GOGET) github.com/gin-gonic/gin
	$(GOGET) github.com/gin-contrib/cors
	$(GOGET) github.com/swaggo/files
	$(GOGET) github.com/swaggo/gin-swagger

docker-build:
	docker build -t webshell-server .

help:
	@echo "Available targets:"
	@echo "  build        - Build the server binary"
	@echo "  build-linux  - Build for Linux (cross-compile)"
	@echo "  build-windows- Build for Windows (cross-compile)"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  swagger      - Generate swagger documentation"
	@echo "  run          - Build and run the server (default: 0.0.0.0:8080)"
	@echo "  run-windows  - Build and run on Windows"
	@echo "  dev          - Run server in development mode"
	@echo "  run-8080     - Run on port 8080"
	@echo "  run-9090     - Run on port 9090"
	@echo "  run-local    - Run on localhost:8080"
	@echo "  run-prod     - Run on 0.0.0.0:80"
	@echo "  install      - Install required dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run HOST=192.168.1.100 PORT=9090"
	@echo "  make dev HOST=127.0.0.1 PORT=8888"