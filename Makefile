PROJECT ?= itmo-high-load-system-1lab
TAG ?= dev
PORT ?= 8080

.PHONY: run test

run:
	go run .

test:
	go test ./...

build:
	docker build -t $(PROJECT):$(TAG) .

docker-run: 
	docker run --rm -p $(PORT):8080 $(PROJECT):$(TAG)