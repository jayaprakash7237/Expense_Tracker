.PHONY: up down run build restart kill-port

# Start PostgreSQL + Run Go (with auto restart + port kill)
up:

	@echo "Starting PostgreSQL..."
	@docker compose up -d postgres
	@echo " Waiting for PostgreSQL..."
# 	@sleep 3
	@echo "PostgreSQL ready"
	@echo "Running Go application..."
	@go run main.go

# Build binary
build:
	@echo "📦 Building Go binary..."
	@go build -o expense-tracker main.go
	@echo "✅ Binary built: ./expense-tracker"

# Run binary (PostgreSQL must be running)
run:
	@echo "🚀 Running binary..."
	@./expense-tracker

# Build + Run
build-run: build run

# Stop PostgreSQL
down:
	@echo "🛑 Stopping PostgreSQL..."
	@docker compose down
	@echo "✅ Done"

# Start PostgreSQL only
db-up:
	@docker compose up -d postgres

# Stop PostgreSQL only
db-down:
	@docker compose down

# Restart everything (down + up)
restart:
	@echo "🔄 Restarting all services..."
	@docker compose down
	@docker compose up -d postgres
	@echo "✅ PostgreSQL restarted"
	@echo "🚀 Running Go application..."
	@go run main.go

# Kill process on port 8080
kill-port:
	@echo "🔍 Killing process on port 8080..."
	@sudo kill -9 $$(sudo lsof -t -i:8080) 2>/dev/null || echo "✅ No process on port 8080"

# Check port status
check-port:
	@echo "🔍 Checking port 8080..."
	@lsof -i :8080 || echo "✅ Port 8080 is free"

# Clean everything
clean:
	@echo "🧹 Cleaning up..."
	@docker compose down -v
	@rm -f expense-tracker
	@echo "✅ Cleanup complete"

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "  make up          - Kill port 8080 + Start PostgreSQL + Run Go"
	@echo "  make down        - Stop PostgreSQL"
	@echo "  make restart     - Restart all services + Run Go"
	@echo "  make build       - Build Go binary"
	@echo "  make run         - Run binary"
	@echo "  make build-run   - Build + Run binary"
	@echo "  make check-port  - Check if port 8080 is in use"
	@echo "  make kill-port   - Kill process on port 8080"
	@echo "  make clean       - Clean everything"