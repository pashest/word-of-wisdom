export GOPROXY=https://proxy.golang.org,direct

.PHONY: build
build: build-server build-client

.PHONY: run
run: run-server run-client

.PHONY: test
test:
	go test -v ./...

.PHONY: install-lint
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint
lint: install-lint
	golangci-lint run ./...

.PHONY: build-server
build-server:
	go build -o bin/server cmd/server/main.go

.PHONY: build-client
build-client:
	go build -o bin/client cmd/client/main.go

.PHONY: run-server
run-server:
	go run cmd/server/main.go

.PHONY: run-client
run-client:
	go run cmd/client/main.go

start:
	docker-compose up --abort-on-container-exit --force-recreate --build server --build client

