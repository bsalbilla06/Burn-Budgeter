# Burn Budgeter API Documentation (Public)

This documentation outlines the public JSON API for Burn Budgeter, a financial observability tool for tracking cloud and AI burn rates.

## Base URL
`http://localhost:8080/v1`

## Authentication
**None.** This API is public. All projects and services are modifiable by any client.

## Standard Error Response
All error responses follow this JSON structure:
```json
{
  "error": "Short error code",
  "message": "Human-readable explanation of the error"
}
```

---

## 1. System

### Health Check
**`GET /health`**
Returns the status of the API and the Supabase database connection.

---

## 2. Projects

### Create a Project
**`POST /projects`**
Initializes a new project.

### Get Project Details
**`GET /projects/{id}`**
Returns project stats, including calculated monthly **burn rate** and **runway**.

### Update Project
**`PATCH /projects/{id}`**
Update project metadata (name, cash on hand, currency).

### Delete Project
**`DELETE /projects/{id}`**
Permanently deletes a project and its stack.

---

## 3. Service Stack

### Add or Update Service in Stack
**`POST /projects/{id}/stack`**
Adds a service to the project stack or updates its quantity.

### Remove Service from Stack
**`DELETE /projects/{id}/stack/{service_id}`**
Removes a specific service entry from the project stack.

---

## 4. AI Analysis

### Analyze Architecture
**`POST /projects/{id}/analyze`**
Parses an `ARCHITECTURE.md` file. **Note:** This will FULLY RESET the project's stack.

### Export Architecture
**`GET /projects/{id}/export-architecture`**
Generates a professional `ARCHITECTURE.md` markdown file from the current stack.

---

## 5. Service Catalog

### List Available Services
**`GET /services?provider={name}`**
Returns the master list of cloud and AI services and pricing.
*   **Query Param:** `provider` (optional) - Filter results by provider name (e.g., `AWS`, `GCP`, `OpenAI`).

### Create a Custom Service
**`POST /services`**
Define a new custom service and price.

### Update Service
**`PATCH /services/{id}`**
Update an existing service's details or pricing.

### Delete Service
**`DELETE /services/{id}`**
Permanently remove a service from the catalog.
