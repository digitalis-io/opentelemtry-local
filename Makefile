# OpenTelemetry Tempo Lab Makefile
# Helper commands for running the OpenTelemetry and Tempo stack

# Default target
.PHONY: help
help:
	@echo "OpenTelemetry Tempo Lab - Available Commands"
	@echo "==========================================="
	@echo "  make up              - Start all services with Go auto-instrumentation"
	@echo "  make up-python       - Start all services with Python auto-instrumentation"
	@echo "  make down            - Stop all services"
	@echo "  make restart         - Restart all services"
	@echo "  make logs            - View logs from all services"
	@echo "  make logs-tempo      - View Tempo logs only"
	@echo "  make logs-grafana    - View Grafana logs only"
	@echo "  make logs-go-auto    - View Go auto-instrumentation logs"
	@echo "  make logs-python     - View Python app logs"
	@echo "  make ps              - Show running containers"
	@echo "  make clean           - Stop services and clean up data"
	@echo ""
	@echo "Application Commands:"
	@echo "  make build-app       - Build the Go test application"
	@echo "  make run-app         - Run Go app with auto-instrumentation"
	@echo "  make run-app-python  - Run Python app locally (no Docker)"
	@echo "  make run-client      - Run the HTTP client test application"
	@echo "  make test-endpoints  - Test all demo server endpoints"
	@echo ""
	@echo "Testing Commands:"
	@echo "  make test-good       - Send a successful request"
	@echo "  make test-bad        - Send a failing request"
	@echo "  make test-admin      - Send an unauthorized request"
	@echo "  make test-health     - Check service health"
	@echo "  make test-all        - Run all endpoint tests"
	@echo ""
	@echo "URLs:"
	@echo "  Grafana:    http://localhost:3000"
	@echo "  Tempo API:  http://localhost:3200"
	@echo "  Demo App:   http://localhost:8080"

# Docker Compose Commands
.PHONY: up
up:
	@echo "Starting OpenTelemetry Tempo stack with Go auto-instrumentation..."
	docker compose up -d
	@echo ""
	@echo "Services started! Access points:"
	@echo "  - Grafana: http://localhost:3000 (anonymous access enabled)"
	@echo "  - Tempo API: http://localhost:3200"
	@echo ""
	@echo "Run 'make logs' to view service logs"

.PHONY: up-python
up-python:
	@echo "Starting OpenTelemetry Tempo stack with Python auto-instrumentation..."
	docker compose -f docker-compose.python.yaml up -d
	@echo ""
	@echo "Services started! Access points:"
	@echo "  - Grafana: http://localhost:3000 (anonymous access enabled)"
	@echo "  - Tempo API: http://localhost:3200"
	@echo "  - Python Demo App: http://localhost:8080"
	@echo ""
	@echo "Run 'make logs-python' to view Python app logs"

.PHONY: down
down:
	@echo "Stopping all services..."
	docker compose down

.PHONY: restart
restart: down up

.PHONY: logs
logs:
	docker compose logs -f

.PHONY: logs-tempo
logs-tempo:
	docker compose logs -f tempo

.PHONY: logs-grafana
logs-grafana:
	docker compose logs -f grafana

.PHONY: logs-go-auto
logs-go-auto:
	docker compose logs -f go-auto

.PHONY: logs-python
logs-python:
	docker compose -f docker-compose.python.yaml logs -f python-app

.PHONY: ps
ps:
	docker compose ps

.PHONY: clean
clean:
	@echo "Stopping services and cleaning up..."
	docker compose down -v
	rm -rf tempo-data/*
	@echo "Cleanup complete!"

# Application Commands
.PHONY: build-app
build-app:
	@echo "Building Go test application..."
	go build -o test-application test-application.go
	@echo "Build complete!"

.PHONY: run-app
run-app: build-app
	@echo "Running Go test application with OpenTelemetry auto-instrumentation..."
	@echo "The go-auto container will automatically instrument the application"
	@echo ""
	@echo "Make sure the stack is running (make up) before testing!"
	@echo "Application will be available at http://localhost:8080"
	./test-application

.PHONY: run-app-python
run-app-python:
	@echo "Running Python test application locally..."
	@echo "NOTE: For auto-instrumentation, use 'make up-python' instead"
	python3 test-application.py

.PHONY: run-client
run-client: build-app
	@echo "Running HTTP client test application..."
	./test-application

# Testing Commands
.PHONY: test-good
test-good:
	@echo "Testing /good endpoint (should return 200)..."
	curl -s -o /dev/null -w "Status: %{http_code}\n" http://localhost:8080/good || echo "Is the application running? Run 'make run-app' first"

.PHONY: test-bad
test-bad:
	@echo "Testing /bad endpoint (should return 500)..."
	curl -s -o /dev/null -w "Status: %{http_code}\n" http://localhost:8080/bad || echo "Is the application running? Run 'make run-app' first"

.PHONY: test-admin
test-admin:
	@echo "Testing /admin endpoint (should return 401)..."
	curl -s -o /dev/null -w "Status: %{http_code}\n" http://localhost:8080/admin || echo "Is the application running? Run 'make run-app' first"

.PHONY: test-health
test-health:
	@echo "Testing /health endpoint..."
	curl -s http://localhost:8080/health | jq . || echo "Is the application running? Run 'make run-app' first"

.PHONY: test-endpoints
test-endpoints:
	@echo "Testing all endpoints..."
	@echo "========================"
	@$(MAKE) test-health
	@echo ""
	@$(MAKE) test-good
	@echo ""
	@$(MAKE) test-bad
	@echo ""
	@$(MAKE) test-admin

.PHONY: test-all
test-all: test-endpoints
	@echo ""
	@echo "All endpoint tests complete!"
	@echo "Check Grafana at http://localhost:3000 to view traces"

# Load testing (requires hey - https://github.com/rakyll/hey)
.PHONY: load-test
load-test:
	@command -v hey >/dev/null 2>&1 || { echo "hey is required but not installed. Install with: go install github.com/rakyll/hey@latest"; exit 1; }
	@echo "Running load test..."
	@echo "Sending 100 requests to each endpoint..."
	hey -n 100 -c 10 http://localhost:8080/good
	hey -n 100 -c 10 http://localhost:8080/bad
	hey -n 100 -c 10 http://localhost:8080/admin
	@echo "Load test complete! Check traces in Grafana"