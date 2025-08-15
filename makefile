.PHONY: init-db start stop clean

# Переменные
DB_NAME := level0
DB_USER := user_wb
DB_PASS := 123
PG_CONTAINER := postgres

# Инициализация БД
init-db:
	goose -dir internal/db/migrate "postgres://postgres:123@localhost:5432/level0?sslmode=disable"

# Запуск всех сервисов
start:
	docker-compose up -d

# Остановка всех сервисов
stop:
	docker-compose down

# Очистка (остановка + удаление томов)
clean:
	docker-compose down -v

# Проверка состояния БД
check-db:
	docker exec -it $(PG_CONTAINER) psql -U $(DB_USER) -d level0 -c "\l"