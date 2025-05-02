
include .env


DB_URL=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable

migrate-up:
	docker compose exec app migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	docker compose exec app migrate -path ./migrations -database "$(DB_URL)" down

migrate-new:
	@read -p "Enter migration name: " name; \
	docker compose exec app migrate create -ext sql -dir ./migrations -seq "$$name"
	sudo chown $$USER ./migrations/*

generate-docs:
	docker compose exec app swag init -g cmd/server/main.go -o cmd/server/docs

.PHONY: test
test:
	go test -v ./... --coverprofile=coverage.out

.PHONY: test-cover
test-cover:
	go tool cover -html=coverage.out

.PHONY: gen-mocks
gen-mocks:
	mockgen -source=internal/services/auth/service.go -destination=internal/services/auth/mocks/auth_service_mocks.go 
	mockgen --source=internal/repositories/repositories.go --destination=internal/repositories/mocks/repositories_mocks.go
	mockgen --source=internal/eventListeners/eventListeners.go --destination=internal/eventListeners/mocks/eventListeners_mocks.go


