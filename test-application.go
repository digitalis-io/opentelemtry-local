package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// User represents a user in our system
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Response represents a standard API response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// simulateDatabase simulates a database call with random latency
func simulateDatabase(ctx context.Context, operation string) error {
	// Random latency between 10-100ms to make traces interesting
	latency := time.Duration(rand.Intn(90)+10) * time.Millisecond
	log.Printf("Database operation: %s (latency: %v)", operation, latency)

	select {
	case <-time.After(latency):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// simulateExternalAPI simulates calling an external API
func simulateExternalAPI(ctx context.Context, endpoint string) (interface{}, error) {
	log.Printf("Calling external API: %s", endpoint)

	// Simulate network latency
	latency := time.Duration(rand.Intn(200)+50) * time.Millisecond

	select {
	case <-time.After(latency):
		// Simulate some data being returned
		return map[string]interface{}{
			"external_data": fmt.Sprintf("Data from %s", endpoint),
			"timestamp":     time.Now().Unix(),
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// goodHandler handles the /good endpoint
func goodHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Printf("Processing good request from %s", r.RemoteAddr)

	// Simulate some business logic with database calls
	if err := simulateDatabase(ctx, "SELECT users WHERE active=true"); err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	// Simulate calling an external service
	externalData, err := simulateExternalAPI(ctx, "https://api.example.com/status")
	if err != nil {
		log.Printf("External API error: %v", err)
		// Continue anyway for demo purposes
	}

	// Create some users data
	users := []User{
		{ID: 1, Name: "Alice Johnson"},
		{ID: 2, Name: "Bob Smith"},
		{ID: 3, Name: "Charlie Brown"},
	}

	response := Response{
		Status:  "success",
		Message: "Request processed successfully",
		Data: map[string]interface{}{
			"users":         users,
			"external_data": externalData,
			"processed_at":  time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

	log.Printf("Successfully processed good request")
}

// badHandler handles the /bad endpoint
func badHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Printf("Processing bad request from %s", r.RemoteAddr)

	// Simulate some processing that leads to an error
	if err := simulateDatabase(ctx, "SELECT * FROM non_existent_table"); err != nil {
		log.Printf("Expected database error: %v", err)
	}

	// Simulate multiple failed operations
	operations := []string{
		"validate_user_permissions",
		"check_rate_limits",
		"process_payment",
	}

	for _, op := range operations {
		log.Printf("Operation failed: %s", op)
		// Add some artificial delay to make traces more interesting
		time.Sleep(time.Duration(rand.Intn(20)+5) * time.Millisecond)
	}

	// Try external API call that will "fail"
	_, err := simulateExternalAPI(ctx, "https://api.example.com/broken-endpoint")
	if err != nil {
		log.Printf("External API call failed as expected: %v", err)
	}

	response := Response{
		Status:  "error",
		Message: "Internal server error occurred",
		Data: map[string]interface{}{
			"error_code": "INTERNAL_ERROR",
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}

	log.Printf("Processed bad request with error response")
}

// adminHandler handles the /admin endpoint
func adminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Printf("Admin access attempted from %s", r.RemoteAddr)

	// Simulate authentication check
	authToken := r.Header.Get("Authorization")
	log.Printf("Checking authorization token: %s", authToken)

	// Simulate database call to check permissions
	if err := simulateDatabase(ctx, "SELECT permissions FROM users WHERE token=?"); err != nil {
		log.Printf("Auth database error: %v", err)
	}

	// Simulate permission validation logic
	operations := []string{
		"validate_token_format",
		"check_token_expiry",
		"verify_admin_permissions",
	}

	for _, op := range operations {
		log.Printf("Auth operation: %s", op)
		time.Sleep(time.Duration(rand.Intn(15)+5) * time.Millisecond)
	}

	// Always return unauthorized for demo purposes
	log.Printf("Authorization failed - insufficient permissions")

	response := Response{
		Status:  "error",
		Message: "Unauthorized access - admin privileges required",
		Data: map[string]interface{}{
			"error_code":    "UNAUTHORIZED",
			"required_role": "admin",
			"timestamp":     time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding unauthorized response: %v", err)
	}

	log.Printf("Rejected admin request - unauthorized")
}

// healthHandler provides a simple health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check from %s", r.RemoteAddr)

	response := Response{
		Status:  "healthy",
		Message: "Service is running",
		Data: map[string]interface{}{
			"uptime":    time.Since(startTime).String(),
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

var startTime time.Time

func main() {
	startTime = time.Now()

	// Seed random number generator for consistent but varied latencies
	rand.Seed(time.Now().UnixNano())

	// Set up routes
	http.HandleFunc("/good", goodHandler)
	http.HandleFunc("/bad", badHandler)
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/health", healthHandler)

	// Root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		response := Response{
			Status:  "success",
			Message: "OpenTelemetry Demo Server",
			Data: map[string]interface{}{
				"endpoints": []string{"/good", "/bad", "/admin", "/health"},
				"version":   "1.0.0",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	port := ":8080"
	log.Printf("Starting server on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET /        - Service info")
	log.Printf("  GET /good    - Returns 200 with success response")
	log.Printf("  GET /bad     - Returns 500 with error response")
	log.Printf("  GET /admin   - Returns 401 unauthorized")
	log.Printf("  GET /health  - Health check endpoint")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
