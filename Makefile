.PHONY: build run test docker-up docker-down clean

# Build the main API binary
build:
	go build -o bin/api cmd/api/main.go

# Run the API locally
run:
	go run cmd/api/main.go

# Run tests
test:
	go test ./... -v

# Run the platform using Docker Compose
docker-up:
	docker compose up --build

# Stop docker containers
docker-down:
	docker compose down

# Clean up binaries and temporary files
clean:
	rm -rf bin/
