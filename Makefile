.PHONY: run test tidy docker-up docker-up-build docker-down docker-logs migrate

run:
	go run ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

migrate:
	go run ./cmd/migrate

docker-up:
	docker compose -f deployments/docker/docker-compose.yml up

docker-up-build:
	docker compose -f deployments/docker/docker-compose.yml up --build

docker-down:
	docker compose -f deployments/docker/docker-compose.yml down -v

docker-logs:
	docker compose -f deployments/docker/docker-compose.yml logs -f
