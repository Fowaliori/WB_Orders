# Демонстрационный сервис для обработки заказов
Демонстрационный сервис с использованием Kafka и PostgreSQL. Принимает заказы через kafka, сохраняет их в PostgreSQL и кэширует в памяти

## Технологии
- Go
- PostgreSQL
- Kafka
- Goose migrations
- Docker

## Работа с проектом
Установить Goose, если не установлен
go install github.com/pressly/goose/v3/cmd/goose@latest
