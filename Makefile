.PHONY: build run test swagger

build:
	go build ./...

run:
	go run cmd/server/server.go -mode=dev

test:
	go test ./... -count=1

swagger:
	swag init -g cmd/server/server.go -o docs --parseDependency --parseInternal
