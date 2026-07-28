# go-password-regis - System Description

## Overview

A user registration service with email verification.

Users register in 2 steps: submit username and password, then verify their email with a code. Once verified, users can log in with email and password through a simple UI.

## Architecture (final state)

```
[Frontend] → HTTP → [Registration Server] → gRPC → [Credential Service]
                        ↓
                    [Temporal Worker]
                    ├── Activity: CreateCredential
                    ├── Activity: SendVerificationEmail
                    └── Signal: EmailVerified → ActivateAccount
```

## Components

| Component | Role |
|-----------|------|
| **Registration Server** | Accepts registration requests, manages prospect lifecycle (pending → verified → active) |
| **Credential Service** | Handles password hashing, credential storage and verification |
| **Temporal Worker** | Orchestrates the registration flow, handles email verification signal, cleans up expired registrations |

## Data

- **prospects** - registration state (id, email, status, created_at, expires_at)
- **credentials** - hashed passwords (id, prospect_id, username, password_hash)

## Phases

### Phase 1 - Go Fundamentals + Web Server

Set up a basic HTTP server and learn core Go concepts:
- Variables, types, constants
- Structs, methods, interfaces
- Slices, maps, pointers
- Error handling (`error`, wrapping, `errors.Is`/`As`)
- Packages and modules (`go mod`)
- Functions, control flow, iteration
- JSON marshalling/unmarshalling

**Deliverable:** A running HTTP server with basic endpoints (e.g. `POST /register`, `POST /login`, `GET /health`). Simple UI for registration and login.
Accepts and returns JSON. Runnable via `go run`. README documents how to build and run.

### Phase 2 - Backend Service + gRPC + Postgres

- Define the business logic as a separate backend service with clear function signatures
- Connect the web server to the backend service via gRPC (protobuf definitions, codegen)
- Web server serves real HTTP endpoints externally, calls backend service internally via gRPC
- Add Postgres for persistence (schema, migrations, queries)
- Add unit tests for business logic and gRPC handlers

**Deliverable:** Two services running together. HTTP → gRPC → Postgres. Unit tests passing. `docker-compose` for Postgres.

### Phase 3 - Temporal

- Replace internal orchestration with Temporal workflows
- Registration flow becomes: Activity `CreateCredential` → Activity `SendVerificationEmail` → Signal `EmailVerified` → Activity `ActivateAccount`
- Add cleanup workflow for expired registrations
- Temporal test suite for workflow unit tests

**Deliverable:** Temporal worker as a separate binary. `docker-compose` includes Temporal server + UI. Registration flow fully orchestrated by Temporal.

### Phase 3.5 (TBD) - ConnectRPC

- Migrate from standard gRPC to ConnectRPC

### Phase 4 - Deployment + CI + Integration Tests

- Dockerfile (multi-stage build) for server and worker
- `docker-compose` for full local environment
- GitHub Actions pipeline: lint → test → build
- Integration/blackbox tests against running services
- README with full setup, run, test, and deploy instructions

**Deliverable:** CI green. Full stack runs via `docker-compose up`. Integration tests pass.
