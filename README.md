# Order Service

Микросервис для хранения и получения заказов с использованием:
- PostgreSQL
- Kafka
- in-memory cache
- Go HTTP API

Проект использовался для анализа производительности Go API:
- profiling,
- benchmarking,
- trace analysis,
- load testing,
- CPU/memory optimization.

---

# Stack

- Go 1.23
- PostgreSQL
- Kafka + Zookeeper
- Docker Compose
- Gorilla Mux
- kafka-go
- net/http/pprof

---

# Run

```bash
make dc-up
```

Проверка API:

```bash
curl http://localhost:8082/order/{order_uid}
```

Логи:

```bash
make dc-log
```

Тесты:

```bash
make test
```

---

# Benchmark

```bash
make bench
```

или:

```bash
go test ./internal/api -bench=. -benchmem
```

Пример:

```text
BenchmarkGetOrderHandler-8          66499 ns/op    59056 B/op    2403 allocs/op
```

---

# Profiling

CPU profile:

```bash
make pprof-cpu
```

или:

```bash
curl -s \
"http://localhost:8082/debug/pprof/profile?seconds=30" \
-o profiles/before/cpu.prof
```

Анализ:

```bash
go tool pprof -top profiles/before/cpu.prof
```

Trace:

```bash
make pprof-trace
go tool trace profiles/before/trace.out
```

---

# Load Testing

```bash
brew install vegeta
```

```bash
make load-test
make load-report
```

---

# Bottlenecks Found

Профилирование показало:
- excessive fmt.Sprintf usage
- excessive allocations
- goroutine overhead
- mutex contention
- inefficient string operations

---

# Optimizations Applied

- removed reflection
- reduced fmt.Sprintf usage
- reduced goroutine fanout
- switched Mutex → RWMutex
- reduced allocations
- optimized string operations

---

# Project Structure

```text
cmd/service        - entrypoint
internal/api       - HTTP handlers
internal/cache     - cache
internal/db        - database
profiles/          - benchmark and pprof results
```