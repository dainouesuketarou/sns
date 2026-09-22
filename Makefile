DB_URL=postgres://dev:dev@localhost:5432/sns?sslmode=disable
MIGRATIONS_DIR=db/migrations

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

psql:
	docker compose exec db psql -U dev -d sns

test:
	DATABASE_URL="$(DB_URL)" go test ./...

.PHONY: migrate-up migrate-down migrate-create psql test
