PROJECT ?= itmo-high-load-system-1lab
TAG ?= dev
PORT ?= 8080

.PHONY: run test build docker-run up down ps logs

run:
	go run .

test:
	go test ./...

build:
	docker build -t $(PROJECT):$(TAG) .

docker-run:
	docker run --rm -p $(PORT):8080 $(PROJECT):$(TAG)

up:
	docker compose up -d --build

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f
