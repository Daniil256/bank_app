# bank_app/Makefile

.PHONY: up down start-services start-infra run-gateway run-auth run-customer

# ============================================
# 1. Поднять ВСЕ контейнеры (БД для auth и customer)
# ============================================
up:
	@echo "Starting infrastructure (databases)..."
	@docker compose -f customer-service/docker-compose.yml up -d
	@docker compose -f auth-service/docker-compose.yml up -d
	@docker compose -f kyc-service/docker-compose.yml up -d
	@docker compose -f document-service/docker-compose.yml up -d
	@echo "Infrastructure started"

# ============================================
# 2. Остановить ВСЕ контейнеры
# ============================================
down:
	@echo "Stopping infrastructure..."
	@cd auth-service && docker-compose down
	@cd customer-service && docker-compose down
	@cd kyc-service && docker-compose down
	@cd document-service && docker-compose down
	@echo "Infrastructure stopped"

# ============================================
# 3. Запустить все Go-сервисы (каждый в своём окне)
#    pause — чтобы окно не закрылось при ошибке
# ============================================
run-services:
	@echo "Starting all services..."
	@start "api-gateway" cmd /c "cd api-gateway && go run ./cmd/gateway & pause"
	@start "auth-service" cmd /c "cd auth-service && go run ./cmd/auth & pause"
	@start "customer-service" cmd /c "cd customer-service && go run ./cmd/customer & pause"
	@start "kyc-service" cmd /c "cd kyc-service && go run ./cmd/kyc & pause"
	@start "document-service" cmd /c "cd document-service && go run ./cmd/document & pause"
	@echo "All services started in windows"

# ============================================
# 4. Поднять инфраструктуру + запустить сервисы
# ============================================
run-all: up run-services
	@echo "Infrastructure + services started"

# ============================================
# 5. Точечный запуск (для отладки)
# ============================================
run-g:
	@start "api-gateway" cmd /c "cd api-gateway && go run ./cmd/gateway & pause"

run-a:
	@start "auth-service" cmd /c "cd auth-service && go run ./cmd/auth & pause"

run-c:
	@start "customer-service" cmd /c "cd customer-service && go run ./cmd/customer & pause"

run-k:
	@start "kyc-service" cmd /c "cd kyc-service && go run ./cmd/kyc & pause"

run-d:
	@start "document-service" cmd /c "cd document-service && go run ./cmd/document & pause"
