# Dockerfile
FROM golang:1.24-alpine AS builder

# Создаем рабочую директорию
WORKDIR /app

COPY go.work ./ 
COPY go.mod go.sum ./
COPY chat-service/go.mod chat-service/go.sum ./chat-service/

# Скачиваем зависимости
RUN go mod tidy

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/chat-server-bin ./chat-service/cmd

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/chat-server-bin .

COPY --from=builder /app/chat-service/config ./config

RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Переключаемся на пользователя app
USER appuser

# Экспортируем порт
EXPOSE 5001

# Запускаем приложение
CMD ["./chat-server-bin"]