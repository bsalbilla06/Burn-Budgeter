# Burn Budgeter API Documentation

This documentation outlines the JSON API for Burn Budgeter, a financial observability tool for tracking cloud and AI burn rates.

## Base URL
`http://localhost:8080/v1`

## Authentication
All protected endpoints require a Bearer token in the `Authorization` header:
`Authorization: Bearer <your_jwt_token>`

### Register
**`POST /auth/register`**
Create a new account.
**Request Body:** `{ "email": "user@example.com", "password": "password123" }`

### Login
**`POST /auth/login`**
Exchange credentials for a JWT.
**Request Body:** `{ "email": "user@example.com", "password": "password123" }`
**Response:** `{ "token": "eyJhbG..." }`

## Standard Error Response
All error responses follow this JSON structure:
```json
{
  "error": "Short error code",
  "message": "Human-readable explanation of the error",
  "details": {}
}
```

---

## 1. Projects

### Create a Project [Auth Required]
**`POST /projects`**

Initializes a new project for the authenticated user.

**Request Body:**
| Field | Type | Description |
| :--- | :--- | :--- |
| `name` | `string` | The name of the project. (Required) |
| `cash_on_hand` | `float` | Current total budget/cash available. (Required) |
| `currency` | `string` | ISO currency code (e.g., "USD"). (Default: "USD") |

**Responses:**
*   **201 Created:** Project successfully initialized.
*   **400 Bad Request:** Missing or malformed JSON.
*   **422 Unprocessable Entity:** Validation failed (e.g., negative `cash_on_hand`).

### Get Project Details [Auth Required]
**`GET /projects/{id}`**

Returns project stats for a project owned by the authenticated user.

**Responses:**
*   **200 OK:** Success.
*   **401 Unauthorized:** Invalid or missing token.
*   **404 Not Found:** Project ID does not exist or belongs to another user.

---

## 2. Service Stack

### Add Service to Stack [Auth Required]
**`POST /projects/{id}/stack`**

Adds a cloud or AI service to the project's infrastructure list.

**Request Body:**
| Field | Type | Description |
| :--- | :--- | :--- |
| `service_id` | `int` | ID of the service from the `/services` list. |
| `quantity` | `float` | Number of units (e.g., 3 instances, 10M tokens). |

**Responses:**
*   **200 OK:** Service added/updated in stack.
*   **400 Bad Request:** Invalid request parameters.
*   **404 Not Found:** Project or Service ID not found.

### Remove Service from Stack [Auth Required]
**`DELETE /projects/{id}/stack/{service_id}`**

Removes a specific service entry from the project stack.

**Responses:**
*   **204 No Content:** Success.
*   **404 Not Found:** Service not found in project stack.

---

## 3. AI Analysis

### Analyze Architecture [Auth Required]
**`POST /projects/{id}/analyze`**

Parses an `ARCHITECTURE.md` file to automatically detect and suggest services.

**Request Body (multipart/form-data):**
| Field | Type | Description |
| :--- | :--- | :--- |
| `architecture` | `file` | The `ARCHITECTURE.md` file to analyze. |

**Example (curl):**
```bash
curl -X POST http://localhost:8080/v1/projects/uuid-1234/analyze 
  -H "Authorization: Bearer <token>" 
  -F "architecture=@ARCHITECTURE.md"
```

**Responses:**
*   **200 OK:** Success with suggestions.
*   **400 Bad Request:** File is missing, too large (>1MB), or empty.
*   **401 Unauthorized:** Invalid or missing token.
*   **422 Unprocessable Entity:** AI could not parse any meaningful services.

---

## 4. Reference Data

### List Available Services [Auth Required]
**`GET /services`**

Returns the master list of supported cloud and AI services.

**Responses:**
*   **200 OK:** Success.
