.PHONY: build run test docker-up docker-down clean swagger

# Build the main API binary
build:
	go build -o bin/api cmd/api/main.go

# Run the API locally
run:
	go run cmd/api/main.go

# Run tests
test:
	go test ./... -v

# Regenerate OpenAPI/Swagger docs (docs/swagger.yaml, docs/docs.go)
swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Run the platform using Docker Compose
docker-up:
	docker compose up --build

# Stop docker containers
docker-down:
	docker compose down

# Clean up binaries and temporary files
clean:
	rm -rf bin/
