.PHONY: help db-up db-down migrate backend-run backend-build frontend-install frontend-dev frontend-build

help:
	@echo "My-Blog development commands"
	@echo "  make db-up            Start PostgreSQL via docker-compose"
	@echo "  make db-down          Stop PostgreSQL"
	@echo "  make migrate          Run database migrations"
	@echo "  make backend-run      Run Go API server"
	@echo "  make backend-build    Build Go binary"
	@echo "  make frontend-install Install frontend dependencies"
	@echo "  make frontend-dev     Start Vite dev server"
	@echo "  make frontend-build   Build frontend for production"

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate:
	cd backend && go run ./cmd/migrate

backend-run:
	cd backend && go run ./cmd/server

backend-build:
	cd backend && go build -o bin/server ./cmd/server

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
