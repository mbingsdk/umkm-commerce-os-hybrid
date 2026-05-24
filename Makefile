.PHONY: db-up db-down db-logs backend-fmt backend-test backend-build backend-run backend-worker backend-migrate frontend-install frontend-dev frontend-build frontend-lint frontend-typecheck

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres

backend-fmt:
	cd backend && go fmt ./...

backend-test:
	cd backend && go test ./...

backend-build:
	cd backend && go build ./...

backend-run:
	cd backend && go run ./cmd/api

backend-worker:
	cd backend && go run ./cmd/worker

backend-migrate:
	cd backend && go run ./cmd/migrate up

backend-seed:
	cd backend && go run ./cmd/seed

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-lint:
	cd frontend && npm run lint

frontend-typecheck:
	cd frontend && npm run typecheck
