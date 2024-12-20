
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