# Messager

Микросервисный мессенджер на Go + PostgreSQL.

## Архитектура

```
own-messager/
├── user-service/      # Регистрация и поиск пользователей (порт 8081)
├── message-service/   # Сообщения и файлы (порт 8082)
├── bff/               # BFF — агрегирует API, long polling, отдаёт фронтенд (порт 8080)
├── frontend/          # HTML/JS фронтенд
└── docker-compose.yml
```

## Запуск

```bash
docker compose up --build
```

Открыть браузер: **http://localhost:8080**

## API (через BFF на порту 8080)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/users` | Регистрация (body: `{"name":"..."}`) |
| GET | `/api/v1/users?q=...` | Поиск пользователей |
| GET | `/api/v1/users/:id` | Получить пользователя |
| POST | `/api/v1/messages` | Отправить сообщение |
| PUT | `/api/v1/messages/:id` | Изменить сообщение (только автор) |
| DELETE | `/api/v1/messages/:id?user_id=...` | Удалить сообщение (только автор) |
| GET | `/api/v1/messages?user_a=&user_b=` | История переписки |
| GET | `/api/v1/poll?user_a=&user_b=&after_id=` | Long polling новых сообщений |
| POST | `/api/v1/files` | Загрузить файл/картинку (multipart form: `file`) |
| GET | `/api/v1/files/:id` | Скачать файл |

## Разработка

Каждый сервис — независимый Go-модуль:

```bash
# user-service
cd user-service && make run

# message-service
cd message-service && make run

# bff
cd bff && make run
```

Переменные окружения — см. `.env.example` в каждом сервисе.

## Миграции

Применяются автоматически при `docker compose up` через контейнеры `migrate-users` и `migrate-messages` (goose).

Ручной запуск:
```bash
cd user-service && make migrate DB_DSN="postgres://messager:messager@localhost:5432/messager_users?sslmode=disable"
cd message-service && make migrate DB_DSN="postgres://messager:messager@localhost:5432/messager_messages?sslmode=disable"
```
