# Skykin Platform

A high-performance SDK platform featuring real-time intent analysis and modular clean architecture.

## Folder Structure

```
skykin-platform/
│
├── cmd/
│   └── api/
│       └── main.go                 # Application Entrypoint
│
├── configs/
│   ├── config.go                   # Configuration Loader
│   ├── env.go                      # Env Helpers
│   └── logger.go                   # Standard Logger
│
├── deployments/
│   ├── docker/
│   │   ├── api.Dockerfile          # Backend Dockerfile
│   │   ├── nginx.conf              # Nginx Gateway
│   │   └── ml.Dockerfile           # ML Microservice Dockerfile
│   └── compose/
│       └── docker-compose.yml      # Multi-container Compose Orchestration
│
├── internal/
│   ├── common/                     # Cross-cutting concerns / Shared utilities
│   │   ├── database/               # Postgres setup and migrations
│   │   ├── middleware/             # HTTP Auth, logging, recovery
│   │   ├── websocket/              # Client hubs and managers
│   │   ├── redis/                  # Cache connectors
│   │   ├── response/               # JSON formatting helpers
│   │   ├── security/               # Hashing & HMAC
│   │   └── utils/                  # Pagination, Validators
│   │
│   ├── users/                      # User Domain (controller, repository, etc.)
│   ├── auth/                       # Auth Domain (validation services)
│   ├── events/                     # Event ingestion and pipelines
│   ├── intents/                    # Intent prediction and ML clients
│   ├── rewards/                    # Reward rules and processing
│   └── websocket/                  # WebSockets routes
│
└── tests/                          # Automated tests suite
```

## Getting Started

### Prerequisites

- Go `1.25.0` or higher
- Docker & Docker Compose
- PostgreSQL

### Running Locally

1. Setup your `.env` file at the root.
2. Run the application:
   ```bash
   make run
   ```

### Running with Docker

Run all services (Database, Backend API, and ML Microservice) using docker-compose:
```bash
make docker-up
```

### API testing (Postman)

Import the collection and environment for **Ad Portal** (frontend), **Developer Portal** (SDK keys), and **SDK** (Flutter):

- [`docs/AD_PORTAL_POSTMAN.md`](docs/AD_PORTAL_POSTMAN.md) — full reference
- [`postman/Skykin-API.postman_collection.json`](postman/Skykin-API.postman_collection.json)
- [`postman/Skykin-Local.postman_environment.json`](postman/Skykin-Local.postman_environment.json)

In Postman: **File → Import** → select both JSON files → activate environment **Skykin — Local**.

**Swagger UI:** `http://localhost:8081/swagger/index.html` (after `make run` or `make docker-up`). Regenerate with `make swagger`.

**Ad portal DB:** Apply `internal/platform/database/migrations/20260603120000_advertisers_campaigns.sql`. Architecture: [`docs/AD_CAMPAIGN_ARCHITECTURE.md`](docs/AD_CAMPAIGN_ARCHITECTURE.md).
