# Burn Budgeter API Documentation

This documentation outlines the JSON API for Burn Budgeter, a financial observability tool for tracking cloud and AI burn rates.

## Base URL
`http://localhost:8080/v1`

## Authentication (Demo Status)
**Current Status: DISABLED for Demo.**
The API is currently configured to bypass token verification and use a hardcoded user account (`9b5940aa-bbf6-40f7-8ce8-30402a8c8737`) for all requests. 

*Planned Implementation:* Protected endpoints will require a Bearer token issued by Supabase Auth in the `Authorization` header.

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
**`GET /health`** (Public)
Returns the status of the API and the Supabase database connection.

---

## 2. Projects

### Create a Project
**`POST /projects`**
Initializes a new project.

**Request Body:**
| Field | Type | Description |
| :--- | :--- | :--- |
| `name` | `string` | The name of the project. (Required) |
| `cash_on_hand` | `float` | Current total budget/cash available. (Required) |
| `currency` | `string` | ISO currency code (e.g., "USD"). (Default: "USD") |

### Get Project Details
**`GET /projects/{id}`**
Returns project stats, including calculated monthly **burn rate** and **runway**.

---

## 3. Service Stack

### Add or Update Service in Stack
**`POST /projects/{id}/stack`**
Adds a cloud or AI service to the project's infrastructure list. If the service already exists in the stack, it updates the quantity.

**Request Body:**
| Field | Type | Description |
| :--- | :--- | :--- |
| `service_id` | `int` | ID of the service from the `/services` list. |
| `quantity` | `float` | Number of units per month (e.g., 730 hours, 10M tokens). |

### Remove Service from Stack
**`DELETE /projects/{id}/stack/{service_id}`**
Removes a specific service entry from the project stack.

---

## 4. AI Analysis

### Analyze Architecture
**`POST /projects/{id}/analyze`**
Parses an `ARCHITECTURE.md` file to automatically detect services. 
**Note:** This will FULLY RESET the project's current tech stack and replace it with the detected services.

**Request Body (multipart/form-data):**
| Field | Type | Description |
| :--- | :--- | :--- |
| `architecture` | `file` | The `ARCHITECTURE.md` file to analyze. |

### Export Architecture
**`GET /projects/{id}/export-architecture`**
Generates a professional `ARCHITECTURE.md` markdown file based on the project's current tech stack using AI.

---

## 5. Reference Data

### List Available Services
**`GET /services`**
Returns the master list of supported cloud and AI services along with their 2026 pricing data.
