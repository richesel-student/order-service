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
   git clone https://github.com/richesel-student/order-service.git
   cd order-service
   ```

2. Собрать и поднять сервисы:
    ```bash
   docker compose down -v --remove-orphans
   docker compose up -d --build
   ```

3. Проверить список контейнеров: 
   ```bash
    docker compose ps
    ```

4. Проверить логи приложения:
    ```bash
    docker logs order-service-app --tail=100
    ```

# Запуск
```bash
    docker compose run --rm producer
```
# Проверка

2. Проверка данных в PostgreSQL
```bash
    docker exec -it order-service-postgres \
    psql -U postgres -d orderservice -c "SELECT order_uid FROM orders LIMIT 5;"
```
3. Получение заказа через API
```bash
    curl http://localhost:8082/order/b563feb7b2b84b6test
```
4. Веб-интерфейс
```bash
http://localhost:8082
```
