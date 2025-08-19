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
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```
Запуск сервисов в docker
```bash
make start
```
Поднятие базы данных
```bash
make init-db
```
Запуск приложения
```bash
go run cmd/app/main.go
```
Запуск генератора 
```bash
go run cmd/producer/main.go
```
Запуск через makefile команды
```bash
#Запуск приложения
make run
#Запуск генератора заказов
make generator
```
