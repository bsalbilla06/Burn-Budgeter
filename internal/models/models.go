package models

import "time"

// Service represents a cloud or AI service
type Service struct {
	ID           int     `json:"id"`
	Provider     string  `json:"provider"`
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	PricePerUnit float64 `json:"price_per_unit"`
}

// Project represents a user's project
type Project struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id,omitempty"` // For future Auth implementation
	Name       string    `json:"name"`
	CashOnHand float64   `json:"cash_on_hand"`
	Currency   string    `json:"currency"`
	CreatedAt  time.Time `json:"created_at"`
	BurnRate   float64   `json:"burn_rate,omitempty"`
	Runway     float64   `json:"runway,omitempty"`
}

// ProjectService represents a service added to a project's stack
type ProjectService struct {
	ProjectID   string  `json:"project_id"`
	ServiceID   int     `json:"service_id"`
	Quantity    float64 `json:"quantity"`
	IsOptimized bool    `json:"is_optimized"`
}

// ErrorResponse represents the standard API error structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
