# Burn Budgeter - Technical Specification (Hackathon MVP)

## 1. Overview
Burn Budgeter is a financial observability API for developers and startups. It provides a real-time "Burn & Runway" calculation by analyzing project stacks across Cloud (AWS/GCP) and AI (Gemini/OpenAI) services. It features an AI-powered `ARCHITECTURE.md` parser to automatically seed infrastructure costs.

## 2. Tech Stack
- **Backend:** Go 1.22+ (Standard library `net/http` for JSON API).
- **Authentication:** Supabase Auth (Managing users and JWT verification).
- **Database:** Supabase (PostgreSQL with Row Level Security).
- **AI Integration:** Google Gemini API (for parsing `ARCHITECTURE.md`).
- **Documentation:** OpenAPI 3.0 (Scalar).
- **Hosting** Supabase / AWS

## 3. Core API Features
1.  **User Authentication:** Managed by Supabase Auth (Sign-up, Login, JWT issuing).
2.  **Project Management:** CRUD endpoints for user-owned projects (Name, Cash on Hand, Currency).
3.  **Service Stack Engine:**
    *   Add/Remove cloud instances (AWS EC2, RDS, S3).
    *   Add/Remove AI models (Gemini, OpenAI, Anthropic).
    *   Support for usage-based units (hours, 1k tokens, GB-months).
4.  **Runway & Burn Analytics:** Calculated fields for monthly burn and "Death Date" estimation.
5.  **AI Architecture Parser:**
    *   Endpoint to POST `ARCHITECTURE.md` content.
    *   Returns a list of detected services and suggested quantities for verification.

## 4. Database Schema (Supabase/Postgres)
- `users` (Managed by Supabase `auth.users`):
    - `id`, `email`, `created_at`.
- `services`: Master list of provider pricing (Hardcoded/Seeded).
    - `id`, `provider`, `name`, `unit`, `price_per_unit`.
- `projects`:
    - `id`, `user_id` (UUID references auth.users.id), `name`, `cash_on_hand`, `currency`, `created_at`.
- `project_services`: Junction table for a project's active stack.
    - `project_id`, `service_id`, `quantity`, `is_optimized`.

## 5. API Endpoints (JSON)
| Method | Route | Description | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/auth/register` | Create a new user account | No |
| `POST` | `/auth/login` | Login and receive JWT | No |
| `POST` | `/projects` | Create a new project | Yes |
| `GET` | `/projects/{id}` | Get project details (Burn & Runway) | Yes |
| `POST` | `/projects/{id}/stack` | Add a service to the stack | Yes |
| `DELETE` | `/projects/{id}/stack/{sid}`| Remove service from stack | Yes |
| `POST` | `/projects/{id}/analyze` | Upload `ARCHITECTURE.md` for autofill | Yes |
| `GET` | `/services` | List available services & pricing | Yes |

## 6. Project Structure
```text
.
├── cmd/
│   └── api/
│       └── main.go          # Entry point
├── internal/
│   ├── handlers/            # JSON API handlers
│   ├── models/              # DB structs & logic
│   ├── database/            # Postgres connection & migrations
│   └── parser/              # Gemini integration for ARCHITECTURE.md
├── api/
│   └── openapi.yaml         # API Documentation
├── scripts/
│   └── seed_pricing.sql     # Initial data for AWS/Gemini
├── Makefile                 # Build & Run commands
└── SPEC.md                  # This file
```

## 7. AI Parsing Logic
*   **The Prompt:** Send `ARCHITECTURE.md` content to Gemini with a system instruction to "extract cloud and AI services mentioned into a JSON array of {provider, service_name, quantity}."
*   **The Mapping:** Match the LLM output against the local `services` database. Return "best guesses" to the user for confirmation.
