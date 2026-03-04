# Social Network - Анкеты пользователей

Простое монолитное приложение социальной сети.

# Требования
- Go 1.21+
- PostgreSQL 12+

# Клонирование репозитория
- git clone <repository-url>
- cd social-network

# Docker
docker-compose up --build -d

# API Endpoints
- /user/register - регистрация
- /login - авторизация
- /user/get/{user_id} - получение анктеты по id-пользователя
- /user/search/?last_name=VALUE_1&first_name=VALUE_2 - поиск анкеты по части фамилии и части имени
- /friend/add/{user_id} - добавление пользователя в список друзей
- /friend/delete/{user_id} - удаление пользователя из списка друзей
- /post/create/ - создание поста
- /post/get/{id} - получение содержимого поста
- /post/delete/{id} - удаление поста
- /post/feed - лента постов
- /post/feed/count - количество постов в ленте пользователя
- /generate/users - генерация данных (пользователи, посты, ленты). Создаются несколько пользователей. Все подписываются на одного пользователя и начинают публиковать посты из файла post.txt.
- /dialog/{user_id}/send - отправка сообщения пользователю user_id
- /dialog/{user_id}/list - получение диалога между двумя пользователями