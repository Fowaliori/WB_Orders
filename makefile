.PHONY: init-db start stop clean run generator


# Инициализация БД
init-db:
	psql -U postgres -h localhost -p 5433 -c "CREATE DATABASE level0;" postgres
	goose -dir internal/db/migrate postgres "postgres://postgres:123@localhost:5433/level0?sslmode=disable" up

# Запуск всех сервисов
start:
	docker-compose up -d

# Остановка всех сервисов
stop:
	docker-compose stop

# Очистка 
clean:
	docker-compose down -v

# Проверка состояния БД
check-db:
	docker exec -it $(PG_CONTAINER) psql -U $(DB_USER) -d level0 -c "\l"

# Запуск основного приложения
run:
	go run cmd/app/main.go

# Запуск генератора данных
generator:
	go run cmd/producer/main.go