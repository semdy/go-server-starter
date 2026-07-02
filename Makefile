.PHONY: build run test swagger install enum generate

build:
	CGO_ENABLED=0 go build -o ./bin/server ./cmd/server

run:
	go run cmd/server/server.go -mode=dev

test:
	go test ./... -count=1

swagger:
	swag init -g cmd/server/server.go -o docs --parseDependency --parseInternal

install:
	go install github.com/abice/go-enum@latest
	go install github.com/swaggo/swag/cmd/swag@latest

enum:
	go generate ./internal/enum/...

generate:
	@if [ -z "$(MODULE)" ]; then echo "Usage: make generate MODULE=product"; exit 1; fi
	./generate.sh $(MODULE)

rebuild_db:
	docker exec -i go-starter-mysql-dev mysql -u root -proot -e "DROP DATABASE IF EXISTS test_db; CREATE DATABASE test_db CHARACTER SET utf8mb4"
