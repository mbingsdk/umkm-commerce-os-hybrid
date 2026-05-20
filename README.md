# UMKM Commerce OS Hybrid

Initial monorepo foundation for **UMKM Commerce OS Hybrid**, a multi-tenant SaaS commerce platform for Indonesian UMKM.

This repository currently contains only the Sprint 0 foundation:

- `backend/` -> Go modular-monolith scaffold
- `frontend/` -> Next.js App Router scaffold
- `docker-compose.yml` -> local PostgreSQL development service
- root tooling for common local commands

Business features such as auth, checkout, POS, payment confirmation, finance, discovery, and admin are intentionally **not implemented yet**.

## Prerequisites

- Go 1.23+
- Node.js 20+
- npm
- Docker Desktop or another Docker Compose-compatible runtime
- `make` if you want to use the convenience commands below

## Local setup

1. Copy the example environment files:

   ```powershell
   Copy-Item backend/.env.example backend/.env
   Copy-Item frontend/.env.example frontend/.env.local
   ```

2. Start PostgreSQL:

   ```powershell
   docker compose up -d postgres
   ```

3. Verify the backend scaffold:

   ```powershell
   cd backend
   go test ./...
   go run ./cmd/api
   ```

4. Install frontend dependencies and start the dev server:

   ```powershell
   cd frontend
   npm install
   npm run dev
   ```

5. Open the frontend at:

   ```txt
   http://localhost:3000
   ```

## Common commands

| Command | Purpose |
|---|---|
| `make db-up` | Start local PostgreSQL |
| `make db-down` | Stop local services |
| `make db-logs` | Follow PostgreSQL logs |
| `make backend-test` | Run Go tests |
| `make backend-build` | Compile backend scaffold |
| `make backend-run` | Run the backend placeholder entrypoint |
| `make frontend-install` | Install frontend dependencies |
| `make frontend-dev` | Start Next.js dev server |
| `make frontend-build` | Build frontend |
| `make frontend-lint` | Run frontend lint |
| `make frontend-typecheck` | Run TypeScript typecheck |

## Repository structure

```txt
.
├─ backend/
│  ├─ cmd/
│  ├─ internal/
│  ├─ migrations/
│  ├─ seeds/
│  └─ tests/
├─ frontend/
│  ├─ app/
│  ├─ components/
│  ├─ features/
│  └─ lib/
├─ docker-compose.yml
├─ Makefile
└─ README.md
```

## Scope boundary for this foundation

The scaffold reserves clear places for later implementation, but it does not yet add:

- tenant/auth logic
- database migrations
- business repositories or services
- checkout/order/POS/payment flows
- finance/discovery/admin behavior

Those belong to the later sprints defined in `PLANS.md` and the project docs.
