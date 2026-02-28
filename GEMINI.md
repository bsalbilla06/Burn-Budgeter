# Burn Budgeter - Gemini CLI Context

This file provides the necessary context and instructions for Gemini CLI to effectively assist with the development and maintenance of the Burn Budgeter project.

## 1. Project Overview
Burn Budgeter is a financial observability API designed for developers and startups. It provides real-time "Burn & Runway" calculations by analyzing project infrastructure stacks across Cloud (AWS/GCP) and AI (Gemini/OpenAI) services.

### Core Features:
- **User Authentication:** JWT-based secure registration and login (Planned).
- **Project Management:** CRUD operations for user-owned projects, including budget and currency tracking.
- **Service Stack Engine:** Manage a project's active infrastructure stack (cloud instances, AI models).
- **Runway & Burn Analytics:** Real-time calculation of monthly burn and estimated "Death Date".
- **AI Architecture Parser:** Automatically detect and suggest services by parsing `ARCHITECTURE.md` files using Google Gemini.

### Tech Stack:
- **Language:** Go 1.22+ (using the standard library `net/http` for JSON API).
- **Database:** PostgreSQL (storing user accounts, projects, tech stacks, and pricing data).
- **AI Integration:** Google Gemini API (for architecture parsing).
- **Authentication:** JWT (JSON Web Tokens).
- **Documentation:** OpenAPI 3.0.

---

## 2. Project Structure
```text
.
├── cmd/
│   └── api/
│       └── main.go          # API Entry point and routing
├── internal/
│   ├── handlers/            # HTTP handlers for JSON API endpoints
│   ├── models/              # Domain models and shared structs
│   ├── middleware/          # JWT and other API middleware
│   ├── database/            # Database connection and migrations
│   └── parser/              # Gemini AI integration logic
├── api/
│   └── openapi.yaml         # API Documentation (OpenAPI)
├── scripts/
│   └── seed_pricing.sql     # Database seeds (pricing data)
├── Makefile                 # Build and development commands
├── SPEC.md                  # Project technical specification
└── API_DOCS.md              # Detailed API endpoint documentation
```

---

## 3. Building and Running
The project uses a `Makefile` for common development tasks.

- **Build the API:** `make build` (outputs binary to `bin/api`)
- **Run the API:** `make run` (runs `cmd/api/main.go`)
- **Run Tests:** `make test`
- **Clean Build Artifacts:** `make clean`

**Default API Port:** `8080` (Base URL: `http://localhost:8080/v1`)

---

## 4. Development Conventions

### Coding Style
- **Standard Library First:** Prioritize using the Go standard library (e.g., `net/http`, `encoding/json`).
- **Enhanced Routing:** Utilize Go 1.22's `http.ServeMux` for path variables and method matching.
- **Surgical Edits:** When modifying code, use the `replace` tool for precise, targeted updates.

### Database Integration
- **Mock Data:** Currently, the handlers in `internal/handlers/handlers.go` use in-memory maps and slices as mock databases.
- **MARK Comments:** Use `// MARK: Need ...` comments to flag areas requiring real database implementation (e.g., SQL queries, transactions).

### API Design
- **Multipart Form Uploads:** The architecture analysis endpoint (`POST /v1/projects/{id}/analyze`) accepts a `multipart/form-data` request with an `architecture` file part.
- **Standardized Error Responses:** All error responses follow the `{ "error": "code", "message": "human-readable" }` format.
- **Stateless Authentication:** All protected endpoints (flagged as `[Auth Required]` in `API_DOCS.md`) will require a Bearer JWT token once authentication is implemented.

### Documentation
- **Keep in Sync:** Ensure `SPEC.md` and `API_DOCS.md` are updated immediately when the API design or database schema changes.

---

## 5. Ongoing Tasks & TODOs
- [ ] Implement JWT Authentication middleware and handlers.
- [ ] Transition from mock data to a real PostgreSQL database.
- [ ] Implement the Google Gemini integration for the architecture parser.
- [ ] Finalize the OpenAPI specification in `api/openapi.yaml`.
