include .env
export
PATH_DB = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATION_DIRS = internal/db/migrations
ENV_FILE=.env
PROD_COMPOSE=docker-compose.prod.yml
NOAPP_COMPOSE=docker-compose.noapp.yml
DEV_COMPOSE=docker-compose.dev.yml

# Import DB
importdb:
	docker exec -i postgres-db psql -U root -d master-golang < ./backupdb-master-golang.sql

# Export DB
exportdb:
	docker exec -i postgres-db pg-dump -U root -d master-golang > ./backupdb-master-golang.sql

# Run Server
server:
	go run ./cmd/api

run-binary:
	./bin/myapp

# build binary
build:
	go build -o bin/myapp ./cmd/api

# generate sqlc
sqlc:
	sqlc generate

# ex: make migrate-create NAME=profiles
migrate-create:
	migrate create -ext sql -dir $(MIGRATION_DIRS) -seq $(NAME)

# run all peding magration:
migrate-up:
	migrate -path $(MIGRATION_DIRS) -database "$(PATH_DB)" up

migrate-down:
	migrate -path $(MIGRATION_DIRS) -database "$(PATH_DB)" down $(STEP)

migrate-force:
	migrate -path $(MIGRATION_DIRS) -database "$(PATH_DB)" force $(VERSION)

migrate-drop:
	migrate -path $(MIGRATION_DIRS) -database "$(PATH_DB)" drop

migrate-goto:
	migrate -path $(MIGRATION_DIRS) -database "$(PATH_DB)" goto $(VERSION)

noapp:
	docker compose -f $(NOAPP_COMPOSE) down
	docker compose -f $(NOAPP_COMPOSE) --env-file $(ENV_FILE) up -d --build

stop-noapp:
	docker compose -f $(NOAPP_COMPOSE) down

dev:
	docker compose -f $(DEV_COMPOSE) down
	docker compose -f $(DEV_COMPOSE) --env-file $(ENV_FILE) up --build

stop-dev:
	docker compose -f $(DEV_COMPOSE) down

prod:
	docker compose -f $(PROD_COMPOSE) down
	docker compose -f $(PROD_COMPOSE) --env-file $(ENV_FILE) up -d --build

stop-prod:
	docker compose -f $(PROD_COMPOSE) down

logs-prod:
	docker compose -f $(PROD_COMPOSE) logs -f --tail=100

bash:
	docker exec -it golang-api /bin/sh

.PHONY: importdb exportdb server migrate-create migrate-up migrate-down migrate-force migrate-drop migrate-goto sqlc build run-binary prod stop-prod logs-prod bash noapp stop-noapp dev stop-dev
