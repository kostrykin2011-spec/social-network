# Dockerfile
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы
COPY go.mod go.sum ./
COPY *.go ./

# Скачиваем зависимости
RUN go mod tidy

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

# Финальный образ
FROM alpine:latest

# Устанавливаем зависимости для runtime
RUN apk --no-cache add ca-certificates

# Создаем пользователя app
RUN addgroup -S app && adduser -S app -G app

# Создаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/main .

# Меняем владельца файлов
RUN chown -R app:app /app

# Переключаемся на пользователя app
USER app

# Экспортируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]