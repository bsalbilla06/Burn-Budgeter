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

// --- PROJECT HANDLERS ---

// CreateProject POST /v1/projects
func CreateProject(w http.ResponseWriter, r *http.Request) {
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
		"INSERT INTO projects (name, cash_on_hand, currency) VALUES ($1, $2, $3) RETURNING id, name, cash_on_hand, currency, created_at",
		req.Name, req.CashOnHand, currency,
	).Scan(&project.ID, &project.Name, &project.CashOnHand, &project.Currency, &project.CreatedAt)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create project: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, project)
}

// GetProject GET /v1/projects/{id}
func GetProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	var project models.Project
	err := database.DB.QueryRow(
		"SELECT id, name, cash_on_hand, currency, created_at FROM projects WHERE id = $1",
		projectID,
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

// UpdateProject PATCH /v1/projects/{id}
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	var req struct {
		Name       *string  `json:"name"`
		CashOnHand *float64 `json:"cash_on_hand"`
		Currency   *string  `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	var project models.Project
	err := database.DB.QueryRow(`
		UPDATE projects 
		SET name = COALESCE($1, name), 
		    cash_on_hand = COALESCE($2, cash_on_hand),
		    currency = COALESCE($3, currency),
		    updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, cash_on_hand, currency, created_at`,
		req.Name, req.CashOnHand, req.Currency, projectID,
	).Scan(&project.ID, &project.Name, &project.CashOnHand, &project.Currency, &project.CreatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Project not found")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// DeleteProject DELETE /v1/projects/{id}
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	res, err := database.DB.Exec("DELETE FROM projects WHERE id = $1", projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to delete project")
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "not_found", "Project not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- SERVICE STACK HANDLERS ---

// AddServiceToStack POST /v1/projects/{id}/stack
func AddServiceToStack(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	// Check if project exists
	var pExists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&pExists)
	if err != nil || !pExists {
		respondError(w, http.StatusNotFound, "not_found", "Project not found")
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
	var exists bool
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
	projectID := r.PathValue("id")
	serviceID := r.PathValue("sid")

	res, err := database.DB.Exec("DELETE FROM project_services WHERE project_id = $1 AND service_id = $2", projectID, serviceID)
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

// --- AI ANALYSIS HANDLERS ---

// AnalyzeArchitecture POST /v1/projects/{id}/analyze
func AnalyzeArchitecture(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	// Check if project exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	if err != nil || !exists {
		respondError(w, http.StatusNotFound, "not_found", "Project not found")
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

	// Call Gemini AI API to parse architecture
	detected, err := parser.ParseArchitecture(r.Context(), content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ai_error", "Failed to analyze architecture: "+err.Error())
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM project_services WHERE project_id = $1", projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to clear existing stack")
		return
	}

	var applied []models.ProjectService
	for _, ds := range detected {
		var serviceID int
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
			_, err = tx.Exec("INSERT INTO project_services (project_id, service_id, quantity) VALUES ($1, $2, $3)", projectID, serviceID, ds.Quantity)
			if err == nil {
				applied = append(applied, models.ProjectService{ProjectID: projectID, ServiceID: serviceID, Quantity: ds.Quantity})
			}
		}
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to commit stack update")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"message": "Stack updated", "updated_services": applied})
}

// ExportArchitecture GET /v1/projects/{id}/export-architecture
func ExportArchitecture(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	// Check if project exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	if err != nil || !exists {
		respondError(w, http.StatusNotFound, "not_found", "Project not found")
		return
	}

	rows, err := database.DB.Query(`
		SELECT s.provider, s.name FROM project_services ps
		JOIN services s ON ps.service_id = s.id
		WHERE ps.project_id = $1`, projectID)
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
		respondError(w, http.StatusUnprocessableEntity, "no_services", "Project has no services to export.")
		return
	}

	markdown, err := parser.GenerateArchitectureMarkdown(r.Context(), serviceNames)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ai_error", "Failed to generate architecture: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(markdown))
}

// --- SERVICE CATALOG HANDLERS ---

// ListServices GET /v1/services
func ListServices(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")

	query := "SELECT id, provider, name, unit, price_per_unit FROM services"
	var params []interface{}

	if provider != "" {
		query += " WHERE provider ILIKE $1"
		params = append(params, provider)
	}

	query += " ORDER BY provider, name"

	rows, err := database.DB.Query(query, params...)
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

// CreateService POST /v1/services
func CreateService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider     string  `json:"provider"`
		Name         string  `json:"name"`
		Unit         string  `json:"unit"`
		PricePerUnit float64 `json:"price_per_unit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	if req.Provider == "" || req.Name == "" || req.Unit == "" {
		respondError(w, http.StatusUnprocessableEntity, "validation_failed", "Provider, name, and unit are required.")
		return
	}

	var s models.Service
	err := database.DB.QueryRow(`
		INSERT INTO services (provider, name, unit, price_per_unit) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, provider, name, unit, price_per_unit`,
		req.Provider, req.Name, req.Unit, req.PricePerUnit,
	).Scan(&s.ID, &s.Provider, &s.Name, &s.Unit, &s.PricePerUnit)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create custom service: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, s)
}

// UpdateService PATCH /v1/services/{id}
func UpdateService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	var req struct {
		Provider     *string  `json:"provider"`
		Name         *string  `json:"name"`
		Unit         *string  `json:"unit"`
		PricePerUnit *float64 `json:"price_per_unit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid JSON body")
		return
	}

	var s models.Service
	err := database.DB.QueryRow(`
		UPDATE services 
		SET provider = COALESCE($1, provider),
		    name = COALESCE($2, name),
		    unit = COALESCE($3, unit),
		    price_per_unit = COALESCE($4, price_per_unit)
		WHERE id = $5
		RETURNING id, provider, name, unit, price_per_unit`,
		req.Provider, req.Name, req.Unit, req.PricePerUnit, serviceID,
	).Scan(&s.ID, &s.Provider, &s.Name, &s.Unit, &s.PricePerUnit)

	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Service not found")
		return
	}

	respondJSON(w, http.StatusOK, s)
}

// DeleteService DELETE /v1/services/{id}
func DeleteService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")

	// Check if service is used in any project stack
	var inUse bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM project_services WHERE service_id = $1)", serviceID).Scan(&inUse)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to check service dependencies")
		return
	}

	if inUse {
		respondError(w, http.StatusConflict, "dependency_error", "Cannot delete service because it is currently used in one or more project stacks.")
		return
	}

	res, err := database.DB.Exec("DELETE FROM services WHERE id = $1", serviceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to delete service")
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "not_found", "Service not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
