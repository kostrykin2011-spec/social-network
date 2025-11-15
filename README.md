# Social Network - Анкеты пользователей

Простое монолитное приложение социальной сети для создания и просмотра анкет.

# Требования
- Go 1.21+
- PostgreSQL 12+

# Клонирование репозитория
- git clone <repository-url>
- cd social-network

# Установка зависимостей

- go mod init social-network
- go mod tidy

# Настройка БД PostgreSQL

- CREATE DATABASE social_network;

# Запуск приложения

- go run cmd/main.go

# API Endpoints
- /user/register
- /login
- /user/get/{user_id}