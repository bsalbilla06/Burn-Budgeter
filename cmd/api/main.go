package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/user/burnbudgeter/internal/handlers"
)

func main() {
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

	port := ":8080"
	fmt.Printf("Burn Budgeter API starting on port %s...\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
