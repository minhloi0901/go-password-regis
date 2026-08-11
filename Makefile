.PHONY: up down build logs logs-server logs-credential logs-postgres migrate-up

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose up --build -d

logs:
	docker compose logs

logs-server:
	docker compose logs -f registration-server 

logs-credential:
	docker compose logs -f credential-service

logs-postgres:
	docker compose logs -f postgres

migrate-up:
	./scripts/migrate.sh