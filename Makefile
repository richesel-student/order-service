# Makefile для сборки, Docker и тестирования Kafka

APP_NAME=order-service
DOCKER_COMPOSE_FILE=docker-compose.yml

# -----------------------------
# Сборка бинарника приложения
# -----------------------------
build:
	@echo ">>> Сборка приложения..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/$(APP_NAME) ./cmd/service

# -----------------------------
# Сборка Docker образа приложения
# -----------------------------
docker-build: build
	@echo ">>> Сборка Docker образа..."
	docker build -t $(APP_NAME) .

# -----------------------------
# Запуск всех сервисов через Docker Compose
# -----------------------------
dc-up:
	@echo ">>> Поднимаем все сервисы..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d --build

# -----------------------------
# Проверяем, что контейнеры успешно запущены
# -----------------------------
dc-ps:
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps

# -----------------------------
# Логи сервисов
# -----------------------------

dc-log:
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

# -----------------------------
# Остановка всех сервисов
# -----------------------------
dc-down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

# -----------------------------
# Перезапуск сервисов
# -----------------------------
dc-restart: dc-down dc-up

# -----------------------------
# Запуск producer контейнера
# -----------------------------
dc-producer-up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) run --rm producer

# -----------------------------
# Запуск тестов 
# -----------------------------
test:
	@echo ">>> Запуск Go тестов..."
	go test -v ./...
# -----------------------------
# Автоматическое тестирование Kafka
# -----------------------------

test-kafka: dc-up
	@echo ">>> Отправка тестового сообщения через producer..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) run --rm producer
	@echo ">>> Проверка логов приложения на получение сообщения..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs --tail=20 app

# -----------------------------
# Удаление всех контейнеров и образов
# -----------------------------
clean:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --rmi all --volumes --remove-orphans
	rm -rf ./bin/*


.PHONY: build docker-build dc-up dc-logs dc-down dc-restart dc-producer-up test-kafka clean
