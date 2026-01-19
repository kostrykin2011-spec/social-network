# Dockerfile
FROM golang:1.24-alpine AS builder

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod tidy

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-service ./cmd

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/chat-service .

COPY --from=builder /app/config ./config

RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Переключаемся на пользователя app
USER appuser

# Экспортируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./chat-service"]