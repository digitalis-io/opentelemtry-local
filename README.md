# OpenTelemetry Auto-Instrumentation with Grafana Tempo Lab

This repository demonstrates how to use OpenTelemetry auto-instrumentation for Go and Python applications with Grafana Tempo for distributed tracing. The lab includes demo applications that generate various types of traces and a complete observability stack.

## Overview

This lab showcases:
- **OpenTelemetry Auto-Instrumentation**: Automatic trace generation without code changes for both Go and Python
- **Grafana Tempo**: Distributed tracing backend for storing and querying traces
- **Demo Applications**: Sample applications in Go and Python to generate realistic traces
- **Grafana Dashboard**: Pre-configured visualization for exploring traces

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│ Go Application  │────▶│ OTel Auto-Instr  │────▶│   Tempo     │
└─────────────────┘     └──────────────────┘     └──────┬──────┘
                                                         │
                                                         ▼
                                                  ┌─────────────┐
                                                  │   Grafana   │
                                                  └─────────────┘
```

## Prerequisites

- Docker and Docker Compose
- Go 1.19 or later (for Go example)
- Python 3.8 or later (for Python example)
- Make (optional, but recommended)
- curl and jq (for testing endpoints)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/digitalis-io/opentelemtry-local.git
cd opentelemtry-local
```

### 2. Start the Stack

#### For Go Application:
```bash
make up
```

#### For Python Application:
```bash
make up-python
```

Or without Make:
```bash
# For Go
docker compose up -d

# For Python
docker compose -f docker-compose.python.yaml up -d
```

This starts:
- **Tempo**: Distributed tracing backend (ports 3200, 4317, 4318, 9411, 14268)
- **Grafana**: Visualization UI (port 3000)
- **Memcached**: Cache for Tempo query performance
- **Auto-Instrumentation**: Automatic instrumentation for your chosen language

### 3. Run the Demo Application

#### Go Application (requires local build):
```bash
# Build the application
make build-app

# Run with auto-instrumentation
make run-app
```

#### Python Application (runs in Docker):
The Python application starts automatically when you run `make up-python`. No additional steps needed!

The demo application will be available at http://localhost:8080

### 4. Generate Some Traces

```bash
# Test all endpoints
make test-all

# Or test individual endpoints
make test-good    # Returns 200 OK
make test-bad     # Returns 500 Error
make test-admin   # Returns 401 Unauthorized
make test-health  # Health check
```

### 5. View Traces in Grafana

1. Open Grafana at http://localhost:3000 (no login required)
2. Navigate to **Explore** in the left sidebar
3. Select **Tempo** as the data source
4. Search for traces using:
   - Service name: `test-application`
   - Trace ID (if you have one)
   - Tags like `http.status_code=500`

## Demo Applications

### 1. Test Server (`test-application.go` / `test-application.py`)

A simple HTTP server (available in both Go and Python) with endpoints that simulate different scenarios:

- **`GET /`** - Service information
- **`GET /good`** - Successful request with database calls and external API simulation
- **`GET /bad`** - Failed request that returns 500 error
- **`GET /admin`** - Unauthorized request that returns 401
- **`GET /health`** - Health check endpoint

Each endpoint generates realistic traces with:
- Simulated database operations
- External API calls
- Variable latencies
- Different error scenarios

### 2. HTTP Client (`app/main.go`)

An HTTP client that makes requests to various external APIs:
- Tests both successful and failing requests
- Demonstrates distributed tracing across services
- Shows how traces propagate through HTTP headers

Run it with:
```bash
make run-client
```

## Available Commands

Run `make help` to see all available commands:

```
Docker Commands:
  make up              - Start all services
  make down            - Stop all services
  make restart         - Restart all services
  make logs            - View logs from all services
  make clean           - Stop services and clean up data

Application Commands:
  make build-app       - Build the test application
  make run-app         - Run test application with auto-instrumentation
  make run-client      - Run the HTTP client test application

Testing Commands:
  make test-good       - Send a successful request
  make test-bad        - Send a failing request
  make test-admin      - Send an unauthorized request
  make test-health     - Check service health
  make test-all        - Run all endpoint tests
```

## How Auto-Instrumentation Works

The OpenTelemetry Go auto-instrumentation uses eBPF to automatically inject tracing code into your Go application without modifying the source code. The `go-auto` service in the Docker Compose file:

1. Monitors for the specified binary (`OTEL_GO_AUTO_TARGET_EXE`)
2. Injects instrumentation at runtime
3. Sends traces to Tempo using OTLP protocol

Configuration is done through environment variables:
- `OTEL_EXPORTER_OTLP_ENDPOINT`: Where to send traces (Tempo)
- `OTEL_SERVICE_NAME`: Name of your service in traces
- `OTEL_PROPAGATORS`: Trace context propagation format

## Exploring Traces

### In Grafana

1. **Search for Traces**:
   - Use the search bar to filter by service name
   - Filter by duration, status code, or custom tags
   - Use TraceQL for advanced queries

2. **Trace View**:
   - See the full request flow
   - Identify bottlenecks and slow operations
   - View tags and logs associated with spans

3. **Service Graph** (if enabled):
   - Visualize service dependencies
   - See request rates and error rates
   - Identify critical paths

### Example TraceQL Queries

```
# Find slow requests (> 100ms)
{ duration > 100ms }

# Find failed requests
{ status.code = 2 }

# Find requests to specific endpoint
{ http.target = "/bad" }

# Complex query
{ service.name = "test-application" && duration > 50ms && http.status_code = 500 }
```

## Troubleshooting

### No Traces Appearing

1. Check if all services are running:
   ```bash
   make ps
   ```

2. Check logs for errors:
   ```bash
   make logs-tempo
   make logs-go-auto
   ```

3. Ensure the application binary name matches `OTEL_GO_AUTO_TARGET_EXE` in docker compose.yaml

### Application Not Starting

1. Make sure to build the application first:
   ```bash
   make build-app
   ```

2. Check that port 8080 is not in use

### Grafana Connection Issues

1. Verify Tempo is running and healthy
2. Check the data source configuration in Grafana
3. Ensure the Tempo URL is correct (`http://tempo:3200`)

## Load Testing

To generate more traces for testing:

```bash
# Install hey if not already installed
go install github.com/rakyll/hey@latest

# Run load test
make load-test
```

This sends 100 concurrent requests to each endpoint.

## Customization

### Using Python Instead of Go

This lab includes both Go and Python versions of the demo application. The Python version (`test-application.py`) provides the same endpoints and functionality as the Go version.

To switch from Go to Python auto-instrumentation:

1. **Use the Python Docker Compose file**:
   ```bash
   make up-python
   # or
   docker compose -f docker-compose.python.yaml up -d
   ```

2. **Key differences**:
   - Python uses `opentelemetry-instrument` command for auto-instrumentation
   - No binary compilation needed - Python runs directly
   - Instrumentation is configured via environment variables in the container

3. **Python-specific environment variables**:
   ```yaml
   - OTEL_SERVICE_NAME=test-application-python
   - OTEL_TRACES_EXPORTER=otlp
   - OTEL_EXPORTER_OTLP_ENDPOINT=http://tempo:4318
   - OTEL_PYTHON_LOG_CORRELATION=true
   ```

### Adding Your Own Application

#### For Go Applications:
1. Update `OTEL_GO_AUTO_TARGET_EXE` in docker-compose.yaml to match your binary name
2. Mount your application directory in the `go-auto` service
3. Configure the OTLP endpoint to send traces to Tempo

#### For Python Applications:
1. Create a new service in docker-compose.python.yaml
2. Install OpenTelemetry packages: `opentelemetry-distro` and `opentelemetry-exporter-otlp`
3. Run with: `opentelemetry-instrument python your-app.py`
4. Set the same environment variables as shown in the Python example

### Tempo Configuration

Edit `tempo.yaml` to:
- Adjust retention periods
- Configure different storage backends
- Enable additional features like metrics generation

### Grafana Configuration

The `data-sources.yml` file configures Tempo as a data source. Modify it to:
- Add authentication if needed
- Configure additional data sources
- Enable streaming for better performance

## Next Steps

1. **Integrate with Your Application**: Apply auto-instrumentation to your own Go or Python services
2. **Add Custom Spans**: Use the OpenTelemetry SDK for manual instrumentation
3. **Set Up Alerting**: Configure alerts based on trace data
4. **Scale Up**: Deploy Tempo in a production environment with proper storage
5. **Correlate with Metrics**: Add Prometheus for metrics alongside traces

## Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [OpenTelemetry Go Auto-Instrumentation](https://github.com/open-telemetry/opentelemetry-go-instrumentation)
- [OpenTelemetry Python Documentation](https://opentelemetry-python.readthedocs.io/)
- [Grafana Explore Traces App](https://grafana.com/docs/grafana/latest/explore/simplified-exploration/traces/)

## License

This project is licensed under the MIT License - see the LICENSE file for details.