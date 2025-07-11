#!/usr/bin/env python3
"""
OpenTelemetry Demo Server - Python Version
A simple HTTP server that generates various types of traces for testing
"""

import json
import logging
import random
import time
from datetime import datetime
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Dict, Any

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global start time for uptime calculation
start_time = datetime.now()


class User:
    """Represents a user in our system"""
    def __init__(self, id: int, name: str):
        self.id = id
        self.name = name
    
    def to_dict(self):
        return {"id": self.id, "name": self.name}


def simulate_database(operation: str) -> None:
    """Simulates a database call with random latency"""
    # Random latency between 10-100ms to make traces interesting
    latency = random.randint(10, 100) / 1000  # Convert to seconds
    logger.info(f"Database operation: {operation} (latency: {latency*1000:.0f}ms)")
    time.sleep(latency)


def simulate_external_api(endpoint: str) -> Dict[str, Any]:
    """Simulates calling an external API"""
    logger.info(f"Calling external API: {endpoint}")
    
    # Simulate network latency
    latency = random.randint(50, 250) / 1000  # Convert to seconds
    time.sleep(latency)
    
    # Simulate some data being returned
    return {
        "external_data": f"Data from {endpoint}",
        "timestamp": int(time.time())
    }


class RequestHandler(BaseHTTPRequestHandler):
    """HTTP request handler for the demo server"""
    
    def log_message(self, format, *args):
        """Override to use our logger"""
        logger.info(f"{self.address_string()} - {format % args}")
    
    def send_json_response(self, status_code: int, response: Dict[str, Any]):
        """Helper to send JSON responses"""
        self.send_response(status_code)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(response).encode('utf-8'))
    
    def do_GET(self):
        """Handle GET requests"""
        if self.path == '/':
            self.handle_root()
        elif self.path == '/good':
            self.handle_good()
        elif self.path == '/bad':
            self.handle_bad()
        elif self.path == '/admin':
            self.handle_admin()
        elif self.path == '/health':
            self.handle_health()
        else:
            self.send_error(404, "Not Found")
    
    def handle_root(self):
        """Root handler - returns service info"""
        response = {
            "status": "success",
            "message": "OpenTelemetry Demo Server (Python)",
            "data": {
                "endpoints": ["/good", "/bad", "/admin", "/health"],
                "version": "1.0.0"
            }
        }
        self.send_json_response(200, response)
    
    def handle_good(self):
        """Handle /good endpoint - successful request with various operations"""
        logger.info(f"Processing good request from {self.address_string()}")
        
        try:
            # Simulate some business logic with database calls
            simulate_database("SELECT users WHERE active=true")
            
            # Simulate calling an external service
            external_data = None
            try:
                external_data = simulate_external_api("https://api.example.com/status")
            except Exception as e:
                logger.error(f"External API error: {e}")
                # Continue anyway for demo purposes
            
            # Create some users data
            users = [
                User(1, "Alice Johnson"),
                User(2, "Bob Smith"),
                User(3, "Charlie Brown")
            ]
            
            response = {
                "status": "success",
                "message": "Request processed successfully",
                "data": {
                    "users": [user.to_dict() for user in users],
                    "external_data": external_data,
                    "processed_at": datetime.now().isoformat()
                }
            }
            
            self.send_json_response(200, response)
            logger.info("Successfully processed good request")
            
        except Exception as e:
            logger.error(f"Error processing good request: {e}")
            self.send_error(503, "Service Unavailable")
    
    def handle_bad(self):
        """Handle /bad endpoint - returns 500 error after some processing"""
        logger.info(f"Processing bad request from {self.address_string()}")
        
        # Simulate some processing that leads to an error
        try:
            simulate_database("SELECT * FROM non_existent_table")
        except Exception as e:
            logger.info(f"Expected database error: {e}")
        
        # Simulate multiple failed operations
        operations = [
            "validate_user_permissions",
            "check_rate_limits",
            "process_payment"
        ]
        
        for op in operations:
            logger.info(f"Operation failed: {op}")
            # Add some artificial delay to make traces more interesting
            time.sleep(random.randint(5, 25) / 1000)
        
        # Try external API call that will "fail"
        try:
            simulate_external_api("https://api.example.com/broken-endpoint")
        except Exception as e:
            logger.info(f"External API call failed as expected: {e}")
        
        response = {
            "status": "error",
            "message": "Internal server error occurred",
            "data": {
                "error_code": "INTERNAL_ERROR",
                "timestamp": datetime.now().isoformat()
            }
        }
        
        self.send_json_response(500, response)
        logger.info("Processed bad request with error response")
    
    def handle_admin(self):
        """Handle /admin endpoint - returns 401 unauthorized"""
        logger.info(f"Admin access attempted from {self.address_string()}")
        
        # Simulate authentication check
        auth_token = self.headers.get('Authorization', '')
        logger.info(f"Checking authorization token: {auth_token}")
        
        # Simulate database call to check permissions
        try:
            simulate_database("SELECT permissions FROM users WHERE token=?")
        except Exception as e:
            logger.error(f"Auth database error: {e}")
        
        # Simulate permission validation logic
        operations = [
            "validate_token_format",
            "check_token_expiry",
            "verify_admin_permissions"
        ]
        
        for op in operations:
            logger.info(f"Auth operation: {op}")
            time.sleep(random.randint(5, 20) / 1000)
        
        # Always return unauthorized for demo purposes
        logger.info("Authorization failed - insufficient permissions")
        
        response = {
            "status": "error",
            "message": "Unauthorized access - admin privileges required",
            "data": {
                "error_code": "UNAUTHORIZED",
                "required_role": "admin",
                "timestamp": datetime.now().isoformat()
            }
        }
        
        self.send_json_response(401, response)
        logger.info("Rejected admin request - unauthorized")
    
    def handle_health(self):
        """Handle /health endpoint - health check"""
        logger.info(f"Health check from {self.address_string()}")
        
        uptime = datetime.now() - start_time
        response = {
            "status": "healthy",
            "message": "Service is running",
            "data": {
                "uptime": str(uptime),
                "timestamp": datetime.now().isoformat()
            }
        }
        
        self.send_json_response(200, response)


def run_server(port: int = 8080):
    """Run the HTTP server"""
    server_address = ('', port)
    httpd = HTTPServer(server_address, RequestHandler)
    
    logger.info(f"Starting server on port {port}")
    logger.info("Available endpoints:")
    logger.info("  GET /        - Service info")
    logger.info("  GET /good    - Returns 200 with success response")
    logger.info("  GET /bad     - Returns 500 with error response")
    logger.info("  GET /admin   - Returns 401 unauthorized")
    logger.info("  GET /health  - Health check endpoint")
    
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        logger.info("Server shutting down...")
        httpd.shutdown()


if __name__ == "__main__":
    # Seed random number generator for consistent but varied latencies
    random.seed()
    
    # Run the server
    run_server()