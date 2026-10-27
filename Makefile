# bank_app/Makefile

.PHONY: up down start-services run-gateway run-auth run-customer

# 1. Поднять все контейнеры (БД, Kafka, Redis и т.д.)
up:
	docker-compose up -d

# 2. Остановить контейнеры
down:
	docker-compose down

# 3. Запустить все Go-сервисы параллельно в фоне
#    (каждый в своей подоболочке, чтобы не блокировать терминал)
start-services:
	@echo "Starting all services..."
	@(cd api-gateway && go run ./cmd/gateway) & \
	 (cd auth-service && go run ./cmd/auth) & \
	 (cd customer-service && go run ./cmd/customer) & \
	 wait

# 4. Точечный запуск (если нужно отладить один сервис)
run-gateway:
	cd api-gateway && go run ./cmd/gateway

run-auth:
	cd auth-service && go run ./cmd/auth

run-customer:
	cd customer-service && go run ./cmd/customer