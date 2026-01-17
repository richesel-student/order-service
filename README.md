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
2. Поднимаем все сервисы:   
   ```bash
   make dc-up
    ```
3. Просмотр логов приложения: 
   ```bash
   make dc-log
    ```
4. Запуск всех Go тестов:
    ```bash
    make test
    ```
5. Тестирование Kafka:
    ```bash
    make test-kafka
    ```
## Быстрая проверка работы

1. Отправка тестовых сообщений в Kafka:
```bash
make dc-producer-up
```

2. Получение заказа через HTTP API:
```bash
curl http://localhost:8082/order/{order_uid}
```

3. Веб-интерфейс сервиса:
```bash
http://localhost:8082
```























