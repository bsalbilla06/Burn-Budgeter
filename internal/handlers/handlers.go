package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/user/burnbudgeter/internal/database"
	"github.com/user/burnbudgeter/internal/models"
	"github.com/user/burnbudgeter/internal/parser"
)

// --- HELPERS ---

// DemoUserID is hardcoded for the demo to bypass Supabase Auth complexity.
const DemoUserID = "9b5940aa-bbf6-40f7-8ce8-30402a8c8737"

func getUserID(r *http.Request) string {
	return DemoUserID
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, errCode, message string) {
	respondJSON(w, status, models.ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}

// --- HANDLERS ---

// CreateProject POST /v1/projects
func CreateProject(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req struct {
		Name       string  `json:"name"`
		CashOnHand float64 `json:"cash_on_hand"`
		Currency   string  `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	if req.Name == "" || req.CashOnHand < 0 {
		respondError(w, http.StatusUnprocessableEntity, "validation_failed", "Project name cannot be empty and cash_on_hand must be non-negative.")
		return
	}

	currency := "USD"
	if req.Currency != "" {
		currency = req.Currency
	}

	var project models.Project
	err := database.DB.QueryRow(
		"INSERT INTO projects (user_id, name, cash_on_hand, currency) VALUES ($1, $2, $3, $4) RETURNING id, name, cash_on_hand, currency, created_at",
		userID, req.Name, req.CashOnHand, currency,
	).Scan(&project.ID, &project.Name, &project.CashOnHand, &project.Currency, &project.CreatedAt)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create project: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, project)
}

// GetProject GET /v1/projects/{id}
func GetProject(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	projectID := r.PathValue("id")

	var project models.Project
	err := database.DB.QueryRow(
		"SELECT id, name, cash_on_hand, currency, created_at FROM projects WHERE id = $1 AND user_id = $2",
		projectID, userID,
	).Scan(&project.ID, &project.Name, &project.CashOnHand, &project.Currency, &project.CreatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Project not found.")
		return
	}

	// Calculate Burn Rate (Monthly)
	var burnRate float64
	err = database.DB.QueryRow(`
		SELECT COALESCE(SUM(ps.quantity * s.price_per_unit), 0)
		FROM project_services ps
		JOIN services s ON ps.service_id = s.id
		WHERE ps.project_id = $1`,
		projectID,
	).Scan(&burnRate)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to calculate burn rate")
		return
	}

	project.BurnRate = burnRate
	if burnRate > 0 {
		project.Runway = project.CashOnHand / burnRate
	} else {
		project.Runway = -1 // Infinite
	}

	respondJSON(w, http.StatusOK, project)
}

// AddServiceToStack POST /v1/projects/{id}/stack
func AddServiceToStack(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	projectID := r.PathValue("id")

	// Check project ownership
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND user_id = $2)", projectID, userID).Scan(&exists)
	if err != nil || !exists {
		respondError(w, http.StatusNotFound, "not_found", "Project not found or unauthorized.")
		return
	}

	var req struct {
		ServiceID int     `json:"service_id"`
		Quantity  float64 `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	// Check if service exists
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM services WHERE id = $1)", req.ServiceID).Scan(&exists)
	if err != nil || !exists {
		respondError(w, http.StatusNotFound, "not_found", "Service not found.")
		return
	}

	// Upsert into project_services
	_, err = database.DB.Exec(`
		INSERT INTO project_services (project_id, service_id, quantity) 
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, service_id) DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = NOW()`,
		projectID, req.ServiceID, req.Quantity,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to add service to stack: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Service added to stack"})
}

// RemoveServiceFromStack DELETE /v1/projects/{id}/stack/{sid}
func RemoveServiceFromStack(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	projectID := r.PathValue("id")
	serviceID := r.PathValue("sid")

	// Verify ownership and delete
	res, err := database.DB.Exec(`
		DELETE FROM project_services 
		WHERE project_id = $1 AND service_id = $2
		AND EXISTS (SELECT 1 FROM projects WHERE id = $1 AND user_id = $3)`,
		projectID, serviceID, userID,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to remove service")
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "not_found", "Service not found in stack.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AnalyzeArchitecture POST /v1/projects/{id}/analyze
func AnalyzeArchitecture(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	projectID := r.PathValue("id")

	// Check project ownership
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND user_id = $2)", projectID, userID).Scan(&exists)
	if err != nil || !exists {
		respondError(w, http.StatusNotFound, "not_found", "Project not found.")
		return
	}

	// Limit request size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "File is too large or invalid form data")
		return
	}

	file, _, err := r.FormFile("architecture")
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Architecture file is missing")
		return
	}
	defer file.Close()

	contentBytes, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to read architecture file")
		return
	}
	content := string(contentBytes)

	if len(content) == 0 {
		respondError(w, http.StatusBadRequest, "bad_request", "Architecture file is empty")
		return
	}

	// Call Gemini AI API to parse architecture
	detected, err := parser.ParseArchitecture(r.Context(), content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ai_error", "Failed to analyze architecture: "+err.Error())
		return
	}

	// Map detected services to database IDs
	tx, err := database.DB.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// 1. Clear existing stack
	_, err = tx.Exec("DELETE FROM project_services WHERE project_id = $1", projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to clear existing stack")
		return
	}

	var applied []models.ProjectService
	for _, ds := range detected {
		var serviceID int
		// Try fuzzy match on name and provider
		query := "SELECT id FROM services WHERE provider ILIKE $1"
		params := []interface{}{"%" + ds.Provider + "%"}

		nameFilter := "%" + ds.Name + "%"
		if ds.Provider == "OpenAI" || ds.Provider == "Anthropic" || ds.Provider == "Gemini" {
			if strings.Contains(strings.ToLower(ds.Name), "pro") {
				nameFilter = "%Pro%Input%"
			} else if strings.Contains(strings.ToLower(ds.Name), "sonnet") {
				nameFilter = "%Sonnet%Input%"
			} else if strings.Contains(strings.ToLower(ds.Name), "opus") {
				nameFilter = "%Opus%Input%"
			} else if strings.Contains(strings.ToLower(ds.Name), "mini") || strings.Contains(strings.ToLower(ds.Name), "haiku") || strings.Contains(strings.ToLower(ds.Name), "flash") {
				nameFilter = "%Mini%Input%"
				if strings.Contains(strings.ToLower(ds.Name), "flash") {
					nameFilter = "%Flash%Input%"
				}
			} else {
				nameFilter = "%Input%"
			}
		}

		query += " AND name ILIKE $2 LIMIT 1"
		params = append(params, nameFilter)

		err := tx.QueryRow(query, params...).Scan(&serviceID)
		if err == nil {
			// 2. Insert new service
			_, err = tx.Exec(
				"INSERT INTO project_services (project_id, service_id, quantity) VALUES ($1, $2, $3)",
				projectID, serviceID, ds.Quantity,
			)
			if err == nil {
				applied = append(applied, models.ProjectService{
					ProjectID: projectID,
					ServiceID: serviceID,
					Quantity:  ds.Quantity,
				})
			}
		}
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to commit stack update")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":          "Stack successfully updated from architecture file",
		"updated_services": applied,
	})
}

// ExportArchitecture GET /v1/projects/{id}/export-architecture
func ExportArchitecture(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	projectID := r.PathValue("id")

	// 1. Fetch current services names/providers
	rows, err := database.DB.Query(`
		SELECT s.provider, s.name 
		FROM project_services ps
		JOIN services s ON ps.service_id = s.id
		JOIN projects p ON ps.project_id = p.id
		WHERE p.id = $1 AND p.user_id = $2`,
		projectID, userID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to fetch project services")
		return
	}
	defer rows.Close()

	var serviceNames []string
	for rows.Next() {
		var provider, name string
		if err := rows.Scan(&provider, &name); err == nil {
			serviceNames = append(serviceNames, fmt.Sprintf("%s %s", provider, name))
		}
	}

	if len(serviceNames) == 0 {
		respondError(w, http.StatusUnprocessableEntity, "no_services", "Project has no services in its stack to export.")
		return
	}

	// 2. Call Gemini to generate markdown
	markdown, err := parser.GenerateArchitectureMarkdown(r.Context(), serviceNames)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ai_error", "Failed to generate architecture markdown: "+err.Error())
		return
	}

	// 3. Return as text/markdown or JSON
	w.Header().Set("Content-Type", "text/markdown")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(markdown))
}

// ListServices GET /v1/services
func ListServices(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, provider, name, unit, price_per_unit FROM services ORDER BY provider, name")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to fetch services")
		return
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Provider, &s.Name, &s.Unit, &s.PricePerUnit); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to scan services")
			return
		}
		services = append(services, s)
	}

	respondJSON(w, http.StatusOK, services)
}
