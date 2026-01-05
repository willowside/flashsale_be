# Stage 1, Builder
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod & go.sum first to optimize by docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# 3 main Entry points
# static linking to ensure compatibility in alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/worker ./cmd/worker/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/dlq_worker ./cmd/dlq_worker/main.go

# Stage 2, Final Image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Taipei

WORKDIR /app

# copied binaries: api, worker, dlq_worker
COPY --from=builder /app/bin/api .
COPY --from=builder /app/bin/worker .
COPY --from=builder /app/bin/dlq_worker .

# copy necessary static resources (Lua scripts and env example)
COPY --from=builder /app/scripts ./scripts
# if have sensitive info in .env, it's better to pass env vars via docker-compose instead of baking into the image
COPY --from=builder /app/.env .

# Expose the application port
EXPOSE 8080

# default to run API server
CMD ["./api"]