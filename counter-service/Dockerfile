# Dockerfile
FROM golang:1.24-alpine AS builder

# Создаем рабочую директорию
WORKDIR /app

COPY go.work ./ 
COPY go.mod go.sum ./
COPY counter-service/go.mod counter-service/go.sum ./counter-service/

RUN go mod tidy

COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/counter-server-bin ./counter-service/cmd

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/counter-server-bin .

COPY --from=builder /app/counter-service/config ./config

RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

USER appuser

# Экспортируем порт
EXPOSE 5001

# Запускаем приложение
CMD ["./counter-server-bin"]