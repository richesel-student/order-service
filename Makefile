# ============================================
# Order Service Makefile
# ============================================

APP_NAME=order-service
DOCKER_COMPOSE_FILE=docker-compose.yml

# ============================================
# Build
# ============================================

build:
	@echo ">>> Building application..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/$(APP_NAME) ./cmd/service

docker-build: build
	@echo ">>> Building docker image..."
	docker build -t $(APP_NAME) .

# ============================================
# Docker Compose
# ============================================

dc-up:
	@echo ">>> Starting services..."
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d --build

dc-ps:
	@echo ">>> Containers status..."
	docker compose -f $(DOCKER_COMPOSE_FILE) ps

dc-log:
	@echo ">>> Streaming logs..."
	docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

dc-down:
	@echo ">>> Stopping services..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down

dc-reset:
	@echo ">>> Full reset..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down -v

dc-restart: dc-down dc-up

# ============================================
# Producer
# ============================================

dc-producer-up:
	@echo ">>> Sending test Kafka message..."
	docker compose -f $(DOCKER_COMPOSE_FILE) run --rm producer

# ============================================
# Tests
# ============================================

test:
	@echo ">>> Running tests..."
	go test -v ./...

# ============================================
# Benchmarks
# ============================================

bench:
	@echo ">>> Running benchmarks..."
	go test ./internal/api -bench=. -benchmem

bench-baseline:
	@echo ">>> Running baseline benchmark..."
	mkdir -p profiles/before
	go test ./internal/api -bench=. -benchmem > profiles/before/benchmark.txt

bench-after:
	@echo ">>> Running optimized benchmark..."
	mkdir -p profiles/after
	go test ./internal/api -bench=. -benchmem > profiles/after/benchmark.txt

# ============================================
# pprof
# ============================================

pprof-cpu:
	@echo ">>> Collecting CPU profile..."
	mkdir -p profiles/before
	curl -s "http://localhost:8082/debug/pprof/profile?seconds=30" -o profiles/before/cpu.prof

pprof-heap:
	@echo ">>> Collecting heap profile..."
	mkdir -p profiles/before
	curl -s "http://localhost:8082/debug/pprof/heap" -o profiles/before/heap.prof

pprof-trace:
	@echo ">>> Collecting trace profile..."
	mkdir -p profiles/before
	curl -s "http://localhost:8082/debug/pprof/trace?seconds=10" -o profiles/before/trace.out

pprof-top:
	@echo ">>> CPU profile analysis..."
	go tool pprof -top profiles/before/cpu.prof

trace:
	@echo ">>> Opening trace viewer..."
	go tool trace profiles/before/trace.out

# ============================================
# Load Testing
# ============================================

load-test:
	@echo ">>> Running vegeta load test..."
	echo "GET http://localhost:8082/order/123" | vegeta attack -rate=200 -duration=30s > results.bin

load-report:
	@echo ">>> Vegeta report..."
	vegeta report results.bin

# ============================================
# Benchstat
# ============================================

benchstat:
	@echo ">>> Comparing benchmarks..."
	benchstat profiles/before/benchmark.txt profiles/after/benchmark.txt

# ============================================
# Kafka test
# ============================================

test-kafka:
	@echo ">>> Sending Kafka test message..."
	docker compose -f $(DOCKER_COMPOSE_FILE) run --rm producer

	@echo ">>> Checking app logs..."
	docker compose -f $(DOCKER_COMPOSE_FILE) logs --tail=20 app

# ============================================
# Cleanup
# ============================================

clean:
	@echo ">>> Cleaning project..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down --rmi all --volumes --remove-orphans

	rm -rf ./bin/*
	rm -rf ./profiles/before/*
	rm -rf ./profiles/after/*
	rm -f results.bin

# ============================================
# PHONY
# ============================================

.PHONY: build docker-build dc-up dc-ps dc-log dc-down dc-reset dc-restart dc-producer-up test bench bench-baseline bench-after pprof-cpu pprof-heap pprof-trace pprof-top trace load-test load-report benchstat test-kafka clean