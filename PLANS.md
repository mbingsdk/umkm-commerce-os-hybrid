# PLANS.md - Current Development Plan

## Current Phase

Sprint 7A Inventory API.

## Current Goal

Implement the tenant-scoped backend inventory API for stock listing, stock movements, manual adjustments, and low-stock threshold updates.

## Sprint 0 Tasks

- Create backend Go project structure.
- Create frontend Next.js project structure.
- Add Docker Compose for PostgreSQL.
- Add `.env.example` files.
- Split database migration SQL into migration files.
- Add README with local setup.
- Add Makefile/scripts for dev commands.

## Sprint 1 Tasks

- Implement Go config loader.
- Implement PostgreSQL connection.
- Implement transaction helper.
- Implement standard API response/error.
- Implement request_id/logger/recover middleware.
- Implement health/version endpoints.
- Ensure frontend can run with base layout and providers.

## Do Not Build Yet

Do not implement these until foundation is ready:

- checkout
- POS
- payment confirmation
- finance
- discovery
- admin panel
- AI features
- marketplace sync
- offline POS
