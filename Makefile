.PHONY: run test tidy migrate seeding docker-up docker-up-build docker-compose-build docker-down docker-logs

run:
	go run ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

migrate:
	go run ./cmd/migrate

seeding:
	go run ./cmd/seeding

docker-up:
	docker compose -f deployments/docker/docker-compose.yml up -d

docker-compose-build:
	docker compose -f deployments/docker/docker-compose.yml up --build -d postgres
	docker compose -f deployments/docker/docker-compose.yml run --rm migrate
	docker compose -f deployments/docker/docker-compose.yml run --rm seed
	docker compose -f deployments/docker/docker-compose.yml up --build -d api

docker-up-build: docker-compose-build

docker-down:
	docker compose -f deployments/docker/docker-compose.yml down -v

docker-logs:
	docker compose -f deployments/docker/docker-compose.yml logs -f
