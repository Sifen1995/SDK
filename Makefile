.PHONY: build run test up down logs ping register-test docker-up docker-down clean swagger

# Default dev workflow: Docker only (API on http://localhost:8081). Do not use `make run` while containers are up.
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f backend

# Quick health check against Docker API
ping:
	curl -sf http://localhost:8081/ping | jq .

# Smoke-test ad portal register (Docker must be running)
register-test:
	@EMAIL="test-$$(date +%s)@example.com"; \
	curl -sf -X POST http://localhost:8081/api/v1/ad-portal/register \
		-H 'Content-Type: application/json' \
		-d "{\"name\":\"Test User\",\"email\":\"$$EMAIL\",\"password\":\"SecurePass1!\"}" | jq .

# Build the main API binary (optional; not used by Docker workflow)
build:
	go build -o bin/api cmd/api/main.go

# Local API on :8081 — conflicts with Docker backend. Use `make up` instead.
run:
	@echo "WARNING: Docker backend uses port 8081. Run 'make down' first, or use 'make up' (recommended)."
	@ss -tln 2>/dev/null | grep -q ':8081 ' && { echo "Port 8081 is in use. Stop Docker: make down"; exit 1; } || true
	go run cmd/api/main.go

# Run tests
test:
	go test ./... -v

# Regenerate OpenAPI/Swagger docs
swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Foreground compose (logs in terminal)
docker-up:
	docker compose up --build

docker-down:
	docker compose down

clean:
	rm -rf bin/
