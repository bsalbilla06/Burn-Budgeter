# Burn Budgeter - Gemini CLI Context

This file provides the necessary context and instructions for Gemini CLI to effectively assist with the development and maintenance of the Burn Budgeter project.

## 1. Project Overview
Burn Budgeter is a financial observability API designed for developers and startups. It provides real-time "Burn & Runway" calculations by analyzing project infrastructure stacks across Cloud (AWS/GCP) and AI (Gemini/OpenAI) services.

### Core Features:
- **Project Management:** CRUD operations for public projects, including budget and currency tracking.
- **Service Stack Engine:** Manage a project's active infrastructure stack (cloud instances, AI models).
- **Custom Services:** Full CRUD for custom cloud or AI services and pricing.
- **Runway & Burn Analytics:** Real-time calculation of monthly burn and estimated "Death Date".
- **AI Architecture Parser:** Automatically reset and update a project's stack by parsing `ARCHITECTURE.md` files using Google Gemini.
- **AI Architecture Exporter:** Generate professional `ARCHITECTURE.md` files from existing project stacks.

### Tech Stack:
- **Language:** Go 1.22+ (using the standard library `net/http` for JSON API).
- **Database:** Supabase (PostgreSQL).
- **AI Integration:** Google Gemini API (for architecture parsing and generation).
- **Documentation:** OpenAPI 3.0 (Scalar).

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
│   ├── database/            # Supabase connection logic
│   └── parser/              # Gemini AI integration logic
├── api/
│   └── openapi.yaml         # API Documentation (OpenAPI)
├── scripts/
│   ├── schema.sql           # Database schema (Public)
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

## 4. Environment Variables
The following variables are required in your `.env` file for Supabase and Gemini integration:
- `SUPABASE_DB_CONN`: PostgreSQL connection string. **Use the "Pooler" string (port 6543) from Supabase settings.**
- `GEMINI_API_KEY`: API key for Google Gemini.

---

## 5. Development Conventions

### Coding Style
- **Standard Library First:** Prioritize using the Go standard library (e.g., `net/http`, `encoding/json`).
- **Enhanced Routing:** Utilize Go 1.22's `http.ServeMux` for path variables and method matching.
- **Surgical Edits:** When modifying code, use the `replace` tool for precise, targeted updates.

### Database Integration
- **Public API:** There are no users or authentication. All data is public and modifiable by any API client.
- **MARK Comments:** Use `// MARK: Need ...` comments to flag areas requiring implementation updates.

### API Design
- **Standardized Error Responses:** All error responses follow the `{ "error": "code", "message": "human-readable" }` format.
- **Scalar Documentation:** Documentation is served at `/docs`.

---

## 6. Ongoing Tasks & TODOs
- [x] Implement AI Architecture Parser.
- [x] Implement AI Architecture Exporter.
- [x] Transition to real Supabase PostgreSQL.
- [x] Enable Custom Service CRUD.
- [ ] Finalize the OpenAPI specification in `api/openapi.yaml`.
