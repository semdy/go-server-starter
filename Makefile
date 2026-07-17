.PHONY: build run serve run test up-dev down-dev up down swagger install enum generate rebuild_db

build:
	CGO_ENABLED=0 go build -o ./bin/server ./cmd/server

serve: build
	./bin/server -mode=prod

run:
	go run cmd/server/server.go -mode=dev

test:
	go test ./... -count=1

up-dev:
	docker compose -f docker-compose.dev.yml up -d

down-dev:
	docker compose -f docker-compose.dev.yml down

up:
	docker compose up -d --build

down:
	docker compose down

swagger:
	swag init -g cmd/server/server.go -o docs --parseDependency --parseInternal

install:
	go install github.com/abice/go-enum@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/air-verse/air@latest

enum:
	go generate ./internal/enum/...

generate:
	@if [ -z "$(MODULE)" ]; then echo "Usage: make generate MODULE=product"; exit 1; fi
	./generate.sh $(MODULE)

rebuild_db:
	docker exec -i go-starter-mysql-dev mysql -u root -proot -e "DROP DATABASE IF EXISTS test_db; CREATE DATABASE test_db CHARACTER SET utf8mb4"
