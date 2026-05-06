# Стадия 1: Сборка приложения
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o worker-service .

# Стадия 2: Финальный образ
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/worker-service .

ENTRYPOINT ["./worker-service"]