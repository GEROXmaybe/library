# RESTful API Система учета книг «Городская Библиотека»

Данный проект представляет собой бэкенд-систему (ядро) для автоматизации учета книг, регистрации читателей и фиксации выдачи/возврата литературы. Проект разработан на языке **Go** в рамках производственной практики.

## Технологический стек
- **Язык программирования:** Go 1.21+
- **База данных:** SQLite (встраиваемая СУБД, не требующая установки отдельного сервера)
- **Драйвер БД:** `modernc.org/sqlite` (чистый Go-драйвер, работающий без CGO компилятора)
- **Роутинг:** Стандартный пакет `net/http`
- **Логирование:** Структурированный JSON-логгер `log/slog`
- **Идентификаторы:** UUIDv4 для уникальных ID книг и пользователей (`github.com/google/uuid`)

---

## Структура проекта

```text
library-api/
├── cmd/
│   └── main.go          # Точка входа, запуск HTTP-сервера и роутинг
├── internal/
│   ├── database.go      # Инициализация SQLite, автоматическое создание таблиц
│   ├── handlers.go      # Бизнес-логика, обработка HTTP-запросов и JSON-ответы
│   └── models.go        # Структуры данных (Book, User, Issue)
├── .env                 # Конфигурационный файл (порт, путь к БД)
├── go.mod               # Зависимости и модули Go
└── library.db           # Локальный файл базы данных SQLite (создается автоматически)

 Чтобы запустить сервер нужно ввести команду "go run cmd/main.go"

 Примеры запросов:
 Добавление книги 
 $body = '{"title": "Преступление и наказание", "author": "Фёдор Достоевский", "isbn": "978-5-17-090632-1", "year": 1866}'
$utf8Body = [System.Text.Encoding]::UTF8.GetBytes($body)
Invoke-RestMethod -Uri "http://localhost:8080/books/create" -Method Post -ContentType "application/json; charset=utf-8" -Body $utf8Body

Регистрация читателя
$body = '{"name": "Иван Иванов", "email": "ivan@example.com"}'
$utf8Body = [System.Text.Encoding]::UTF8.GetBytes($body)
Invoke-RestMethod -Uri "http://localhost:8080/users" -Method Post -ContentType "application/json; charset=utf-8" -Body $utf8Body

