Ты — senior Go engineer и performance engineer.

Нужно создать учебный проект по оптимизации Go HTTP API.

# Цель

Взять простой HTTP API, создать нагрузку, собрать профили CPU/Memory/Trace и провести несколько итераций оптимизации.

Финальный результат должен быть оформлен как полноценный GitHub-проект с:
- историей коммитов,
- профилировкой,
- benchmark-результатами,
- README,
- описанием оптимизаций и их эффекта.

# Технологии

Использовать:
- Go 1.22+
- net/http
- net/http/pprof
- pprof
- go test -bench
- benchstat
- go tool trace

Дополнительно можно:
- vegeta или hey для нагрузки
- Makefile
- Dockerfile (опционально)

# Что нужно сделать

## 1. Создать API

Сделай небольшой HTTP API.

Пример:
- GET /users
- GET /users/{id}
- POST /users
- GET /search?q=

Важно:
- API должно быть намеренно НЕоптимальным в первой версии.
- Использовать in-memory storage.
- Добавить JSON serialization.
- Добавить искусственно дорогие операции:
  - лишние аллокации,
  - копирование данных,
  - inefficient string operations,
  - mutex contention,
  - неоптимальные map/slice usage,
  - fmt.Sprintf в hot path,
  - reflection где не нужно,
  - repeated JSON marshal/unmarshal,
  - excessive goroutines.

Нужно чтобы было что оптимизировать.

---

## 2. Подключить pprof

Добавить:
- import _ "net/http/pprof"

Поднять profiling endpoints:
- /debug/pprof/

README должен содержать команды:
- go tool pprof
- top
- list
- web

Примеры:
```bash
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
go tool pprof -http=:8081 cpu.prof