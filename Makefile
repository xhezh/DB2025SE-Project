.PHONY: help setup db-create db-drop db-schema db-seed db-reset run test clean

help: ## Показать доступные команды
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## Установить зависимости Go
	go mod download
	go mod tidy

db-create: ## Создать базу данных
	createdb coworking_db || echo "База данных уже существует"

db-drop: ## Удалить базу данных
	dropdb coworking_db || echo "База данных не существует"

db-schema: ## Применить схему БД
	psql -d coworking_db -f migrations/schema.sql

db-seed: ## Загрузить тестовые данные
	psql -d coworking_db -f migrations/seed.sql

db-reset: db-drop db-create db-schema db-seed ## Пересоздать БД с нуля

run: ## Запустить приложение
	go run cmd/api/main.go

build: ## Собрать бинарник
	go build -o bin/coworking-booking cmd/api/main.go

test: ## Запустить тесты (если есть)
	go test -v ./...

clean: ## Очистить временные файлы
	rm -rf bin/
	go clean

query-test: ## Проверить SQL запросы
	psql -d coworking_db -f migrations/queries.sql

# Быстрый запуск всего проекта
quickstart: setup db-reset run ## Быстрый старт: установка + БД + запуск
