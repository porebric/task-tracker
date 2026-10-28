APP := task-tracker
TEST_MYSQL_DSN ?= root:root@tcp(127.0.0.1:3306)/task_tracker_test
DB_DSN ?= tracker:tracker@tcp(127.0.0.1:3306)/task_tracker
REDIS_ADDR ?= 127.0.0.1:6379
JWT_SECRET ?= local-dev-secret

.PHONY: build run test test-integration lint tidy up down logs

build:
	CGO_ENABLED=0 go build -trimpath -o bin/$(APP) ./cmd/$(APP)

run:
	SERVICE_NAME=$(APP) DB_DSN="$(DB_DSN)" REDIS_ADDR="$(REDIS_ADDR)" JWT_SECRET="$(JWT_SECRET)" go run ./cmd/$(APP)

test:
	go test ./...

test-integration:
	@test -f .env || cp .env.example .env
	docker compose up -d --wait mysql
	TEST_MYSQL_DSN="$(TEST_MYSQL_DSN)" go test -count=1 -run Integration ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

up:
	@test -f .env || cp .env.example .env
	docker compose up --build -d --wait

down:
	docker compose down

logs:
	docker compose logs -f app
