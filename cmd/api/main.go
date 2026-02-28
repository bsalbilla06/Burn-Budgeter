package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/user/burnbudgeter/internal/database"
	"github.com/user/burnbudgeter/internal/handlers"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Initialize Database
	database.Connect()
	defer database.DB.Close()

	mux := http.NewServeMux()

	// Projects
	mux.HandleFunc("POST /v1/projects", handlers.CreateProject)
	mux.HandleFunc("GET /v1/projects/{id}", handlers.GetProject)
	
	// Service Stack
	mux.HandleFunc("POST /v1/projects/{id}/stack", handlers.AddServiceToStack)
	mux.HandleFunc("DELETE /v1/projects/{id}/stack/{sid}", handlers.RemoveServiceFromStack)
	
	// AI Analysis
	mux.HandleFunc("POST /v1/projects/{id}/analyze", handlers.AnalyzeArchitecture)
	
	// Reference Data
	mux.HandleFunc("GET /v1/services", handlers.ListServices)

	// Documentation (Scalar)
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		html := `<!doctype html>
<html>
  <head>
    <title>Burn Budgeter API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/openapi.yaml"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Serve openapi.yaml for Scalar
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "api/openapi.yaml")
	})

	// Health Check (Public)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		err := database.DB.Ping()
		if err != nil {
			http.Error(w, `{"status": "down", "database": "error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "up", "database": "connected"}`))
	})

	port := ":8080"
	fmt.Printf("Burn Budgeter API starting on port %s (AUTH DISABLED)...\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
