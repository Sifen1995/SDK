package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"skykin-platform/configs"
	"skykin-platform/internal/auth/controller"
	"skykin-platform/internal/auth/dto"
	"skykin-platform/internal/auth/repository"
	"skykin-platform/internal/auth/routes"
	"skykin-platform/internal/auth/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestDB configures a local transaction loop so tests don't pollute production data
func SetupTestDB(t *testing.T) *gorm.DB {
	// Connect to your local containerized testing schema instance
	dsn := "postgres://skykin_user:password@localhost:5435/skykin_db?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open integration test DB session: %v", err)
	}
	return db
}

func TestRegisterDeveloperEndpoint(t *testing.T) {
	// Set Gin framework engine to quiet execution context
	gin.SetMode(gin.TestMode)
	r := gin.New()

	db := SetupTestDB(t)

	// Begin a temporary database transaction that we will roll back automatically at the end
	tx := db.Begin()
	defer tx.Rollback()

	cfg := &configs.Config{JwtSecretKey: "test_secret_key"}

	// Spin up real application module layers running against the temporary transaction block
	repo := repository.NewAuthRepository(tx, cfg)
	serv := service.NewAuthService(repo, cfg)
	authCtrl := controller.NewAuthController(serv)

	routes.RegisterAuthRoutes(r, authCtrl)

	// Compose validation payload to trigger our structural DTO tags
	signupPayload := dto.DeveloperRegisterRequest{
		Name:     "Test Sifen",
		Email:    "test_sifen@skykin.io",
		Password: "short", // This should trigger our DTO validation error (min=8 constraint!)
	}

	body, _ := json.Marshal(signupPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/portal/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	respRecorder := httptest.NewRecorder()

	// Execute the HTTP Request down the active Gin route tree
	r.ServeHTTP(respRecorder, req)

	// Assert that validation interceptors successfully intercepted the short password
	assert.Equal(t, http.StatusBadRequest, respRecorder.Code)
	assert.Contains(t, respRecorder.Body.String(), "Validation failed")
}
