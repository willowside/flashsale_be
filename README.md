## High-Concurrency Flash Sale System ##
  This is a Flash Sale system built with Golang. Not just a simple CRUD, this project focus on how to handle high concurrency, data consistency, and observability.

### Key Improvements & Progress ###
  I have updated the system architecture from a simple version to a distributed containerized version. Here are the key things I did:
  * Fixed Panic & Nil Pointer: Added recovery middleware and nil check in order_handler.go so the API won't crash when result is nil.
  * Refactor to Gatekeeper Pattern: Changed the logic from "pre-deduct in Redis" to Gatekeeper Pattern. Now Lua script only checks stock and user purchase record but doesn't DECR. This fixed the issue where Redis stock becomes 0 but DB still has stock.
  * Worker Scaling: Containerized the app and used docker-compose to scale to 3~10 workers. This solved the MQ pending message issue (was 2.9w+ messages stuck) and improved processing speed.
  * DLQ & Compensator: Added Dead Letter Queue and a Compensator worker. When force-fail happens, it will catch the error and do IncrStock in Redis to recover the stock.

### System Architecture ###
  Backend: Golang (Gin)
  DB: PostgreSQL (pgxpool)
  Cache: Redis (Lua Scripting)
  MQ: RabbitMQ (DLQ logic)
  Observability: Grafana, Prometheus, cAdvisor (to see worker CPU)

### 📊 Final K6 Test Result (Showcase) ###
#### Below is the result of 200 VUs test after I scaled the workers: (k6 screenshot) ####
* P95 Latency: 11.83ms (very fast because of Gatekeeper pattern)
* HTTP Req Failed: ~14% (Expected 409/429, no 500 error)
* MQ Result: Handled 4.5w+ requests. Even when MQ is piling up, DB is safe because of traffic shaving.

### Project Structure ###
```plaintext
.
├── cmd/
│   ├── api/            # API entry
│   ├── worker/         # Order worker (can scale)
│   └── dlq_worker/     # DLQ recovery worker
├── internal/
│   ├── handler/        # Handler with nil-pointer fix
│   ├── service/        # Business logic with retry
│   ├── repository/     # Postgres & Redis implementation
│   └── worker/         # Order Processor core logic
├── scripts/            # Lua scripts (Gatekeeper & Finalize)
├── k6/                 # Load test scripts
└── docker-compose.yaml # Docker config with 3 replicas
```
### TODO / Next Steps ###
  * Advanced Error Handling: Implement error classification to distinguish between Transient Errors (e.g., temporary DB downtime) and Permanent Errors (e.g., malformed poison messages). This will allow the system to safely requeue messages for retry or discard invalid data to prevent infinite loops.
  * Local Cache: Add BigCache to reduce Redis load and network overhead.
  * Distributed Tracing: Use Jaeger to observe the whole request lifecycle from API to MQ and Workers.
  * Dynamic Rate Limiter: Automatically adjust the token bucket rate based on system CPU and memory usage.
