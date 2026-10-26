PROJECT ?= itmo-high-load-system-1lab
TAG ?= dev
PORT ?= 8080

.PHONY: app-run app-test docker-build docker-run docker-init docker-config docker-up docker-down docker-ps docker-logs docker-health load-logs load-start load-stop load-smoke

app-run:
	go run ./cmd/server

app-test:
	go test ./...

docker-build:
	docker build -t $(PROJECT):$(TAG) .

docker-run:
	docker run --rm -p $(PORT):8080 $(PROJECT):$(TAG)

docker-init:
	@if [ -e .env ]; then \
		echo '.env уже существует, файл оставлен без изменений.'; \
	else \
		cp .env.example .env && \
		echo 'создан .env. Укажите GRAFANA_ADMIN_PASSWORD, затем выполните make docker-up.'; \
	fi

docker-config:
	docker compose config --quiet

docker-up: docker-config
	docker compose up -d --build

docker-down:
	docker compose down

docker-ps:
	docker compose ps -a

docker-logs:
	docker compose logs --tail=100 -f

docker-health:
	curl --fail --silent --show-error --output /dev/null --write-out 'pod 1: HTTP %{http_code}\n' http://localhost:8080/health
	curl --fail --silent --show-error --output /dev/null --write-out 'pod 2: HTTP %{http_code}\n' http://localhost:8081/health

load-logs:
	docker compose logs --tail=100 -f load

load-start: docker-config
	docker compose up -d --build load

load-stop:
	docker compose stop load

load-smoke: docker-config load-stop
	docker compose run --rm --build \
		-e START_RPS=1 -e RAMP_DURATION=5s -e HOLD_DURATION=15s \
		--entrypoint pandora load /load/load.yaml
