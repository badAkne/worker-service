# Стадия 1: Сборка приложения
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Стадия 2: Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/main .

# Порт по умолчанию (из APP_PROCESSOR_WEB_SERVER_LISTEN_PORT)
EXPOSE 9000

ENTRYPOINT ["./main"]
CMD ["web-server"]
