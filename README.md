# Order Service

Микросервис для хранения и получения заказов.  
Использует **Postgres** для хранения, **Kafka** для приёма новых заказов и **in-memory cache** для быстрых ответов.

---

## Стек

- Go 1.23
- Kafka + Zookeeper
- PostgreSQL
- Docker Compose
- Gorilla Mux (HTTP роутинг)
- kafka-go (работа с Kafka)

---

## Как запустить

1. Склонировать репозиторий:
   ```bash
   git clone https://github.com/<your-repo>/order-service.git
   cd order-service

2. Собрать и поднять сервисы:
   docker compose down -v --remove-orphans
   docker compose up -d --build

3. Проверить список контейнеров: 
    docker compose ps

4. Проверить логи приложения:
    docker logs order-service-app --tail=100

# Запуск
    docker compose run --rm producer

# Проверка

    docker exec -it order-service-postgres \
    psql -U postgres -d orderservice -c "SELECT order_uid FROM orders LIMIT 5;"
