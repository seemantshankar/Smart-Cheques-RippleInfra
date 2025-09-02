package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/config"
	"github.com/smart-payment-infrastructure/internal/handlers"
	"github.com/smart-payment-infrastructure/internal/middleware"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/auth"

	_ "github.com/lib/pq"
)

// IntegrationTestSuite holds the test dependencies
type IntegrationTestSuite struct {
	db                      *sql.DB
	router                  *gin.Engine
	userRepo                *repository.UserRepository
	enterpriseRepo          *repository.EnterpriseRepository
	walletRepo              *repository.WalletRepository
	authService             *services.AuthService
	enterpriseService       *services.EnterpriseService
	walletService           *services.WalletService
	walletMonitoringService *services.WalletMonitoringService
	jwtService              *auth.JWTService
	testUser                *models.User
	testEnterprise          *models.Enterprise
}

func setupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Load environment variables from .env file
	if err := godotenv.Load("../../.env"); err != nil {
		t.Logf("No .env file found, using system environment variables: %v", err)
	}

	// Load test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			PostgresURL: getTestPostgresURL(),
		},
		JWT: config.JWTConfig{
			SecretKey:            "test-secret-key-32-characters-12",
			AccessTokenDuration:  "15m",
			RefreshTokenDuration: "168h",
		},
		XRPL: config.XRPLConfig{
			NetworkURL: "https://s.altnet.rippletest.net:51234",
			TestNet:    true,
		},
	}

	// Connect to test database using environment variables
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	
	connStr := fmt.Sprintf("postgres://user:password@%s:5432/smart_payment?sslmode=disable", dbHost)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	// Clean up any existing test data
	cleanupTestData(t, db)

	// Parse JWT durations
	accessTokenDuration, err := time.ParseDuration(cfg.JWT.AccessTokenDuration)
	require.NoError(t, err)
	refreshTokenDuration, err := time.ParseDuration(cfg.JWT.RefreshTokenDuration)
	require.NoError(t, err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	enterpriseRepo := repository.NewEnterpriseRepository(db)
	walletRepo := repository.NewWalletRepository(db)

	// Initialize XRPL service
	xrplService := services.NewXRPLService(services.XRPLConfig{
		NetworkURL: cfg.XRPL.NetworkURL,
		TestNet:    cfg.XRPL.TestNet,
	})
	require.NoError(t, xrplService.Initialize())

	// Initialize services
	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, accessTokenDuration, refreshTokenDuration)
	authService := services.NewAuthService(userRepo, jwtService)

	walletService, err := services.NewWalletService(
		walletRepo,
		enterpriseRepo,
		xrplService,
		services.WalletServiceConfig{
			EncryptionKey: cfg.JWT.SecretKey,
		},
	)
	require.NoError(t, err)

	enterpriseService := services.NewEnterpriseService(enterpriseRepo, walletService)
	walletMonitoringService := services.NewWalletMonitoringService(walletRepo, xrplService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	enterpriseHandler := handlers.NewEnterpriseHandler(enterpriseService)
	walletHandler := handlers.NewWalletHandler(walletService, walletMonitoringService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Authentication endpoints
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected endpoints
	protected := router.Group("/api")
	protected.Use(handlers.AuthMiddleware(authService))
	{
		protected.POST("/enterprises",
			middleware.RequirePermission(models.PermissionCreateEnterprise),
			enterpriseHandler.RegisterEnterprise)

		protected.GET("/enterprises/:id",
			middleware.RequirePermission(models.PermissionReadEnterprise),
			enterpriseHandler.GetEnterprise)

		protected.POST("/wallets",
			middleware.RequirePermission(models.PermissionCreateEnterprise),
			walletHandler.CreateWallet)

		protected.GET("/wallets/:id",
			middleware.RequirePermission(models.PermissionReadEnterprise),
			walletHandler.GetWallet)

		protected.PUT("/wallets/:id/activate",
			middleware.RequirePermission(models.PermissionApproveKYB),
			walletHandler.ActivateWallet)

		protected.PUT("/wallets/:id/whitelist",
			middleware.RequirePermission(models.PermissionApproveKYB),
			walletHandler.WhitelistWallet)

		protected.GET("/wallets/health-status",
			middleware.RequirePermission(models.PermissionViewAuditLogs),
			walletHandler.GetWalletHealthStatus)
	}

	suite := &IntegrationTestSuite{
		db:                      db,
		router:                  router,
		userRepo:                userRepo,
		enterpriseRepo:          enterpriseRepo,
		walletRepo:              walletRepo,
		authService:             authService,
		enterpriseService:       enterpriseService,
		walletService:           walletService,
		walletMonitoringService: walletMonitoringService,
		jwtService:              jwtService,
	}

	// Create test user and enterprise
	suite.createTestUserAndEnterprise(t)

	return suite
}

func (suite *IntegrationTestSuite) createTestUserAndEnterprise(t *testing.T) {
	// Create test user
	userReq := &models.UserRegistrationRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      "admin",
	}

	user, err := suite.authService.RegisterUser(userReq)
	require.NoError(t, err)
	suite.testUser = user

	// Create test enterprise
	enterpriseReq := &models.EnterpriseRegistrationRequest{
		LegalName:          "Test Enterprise Ltd",
		RegistrationNumber: "TEST123456",
		TaxID:              "TAX123456",
		Jurisdiction:       "US",
		BusinessType:       "Corporation",
		Industry:           "Technology",
		Email:              "contact@testenterprise.com",
		Phone:              "+1234567890",
		Address: models.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		AuthorizedRepresentative: models.AuthorizedRepresentativeRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@testenterprise.com",
			Phone:     "+1234567890",
			Position:  "CEO",
		},
	}

	enterprise, err := suite.enterpriseService.RegisterEnterprise(enterpriseReq)
	require.NoError(t, err)
	suite.testEnterprise = enterprise
}

func (suite *IntegrationTestSuite) getAuthToken(t *testing.T) string {
	accessToken, err := suite.jwtService.GenerateAccessToken(suite.testUser.ID, suite.testUser.Email, suite.testUser.Role, suite.testUser.EnterpriseID)
	require.NoError(t, err)
	return accessToken
}

func (suite *IntegrationTestSuite) cleanup() {
	cleanupTestData(nil, suite.db)
	suite.db.Close()
}

func TestWalletProvisioningIntegration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	t.Run("Enterprise Registration Automatically Creates Wallet", func(t *testing.T) {
		// The wallet should have been created automatically during enterprise registration
		wallets, err := suite.walletService.GetWalletsForEnterprise(suite.testEnterprise.ID)
		require.NoError(t, err)
		assert.Len(t, wallets, 1)

		wallet := wallets[0]
		assert.Equal(t, suite.testEnterprise.ID, wallet.EnterpriseID)
		assert.Equal(t, models.WalletStatusPending, wallet.Status)
		assert.False(t, wallet.IsWhitelisted)
		assert.Equal(t, "testnet", wallet.NetworkType)
		assert.NotEmpty(t, wallet.Address)
		assert.NotEmpty(t, wallet.PublicKey)
	})

	t.Run("Manual Wallet Creation via API", func(t *testing.T) {
		token := suite.getAuthToken(t)

		// Create wallet request
		walletReq := models.WalletCreateRequest{
			EnterpriseID: suite.testEnterprise.ID,
			NetworkType:  "mainnet",
		}

		reqBody, _ := json.Marshal(walletReq)
		req := httptest.NewRequest("POST", "/api/wallets", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Wallet created successfully", response["message"])
		walletData := response["wallet"].(map[string]interface{})
		assert.Equal(t, suite.testEnterprise.ID.String(), walletData["enterprise_id"])
		assert.Equal(t, "mainnet", walletData["network_type"])
		assert.Equal(t, "pending", walletData["status"])
	})

	t.Run("Wallet Activation Workflow", func(t *testing.T) {
		// Get the pending wallet
		wallets, err := suite.walletService.GetWalletsForEnterprise(suite.testEnterprise.ID)
		require.NoError(t, err)

		var pendingWallet *models.WalletResponse
		for _, wallet := range wallets {
			if wallet.Status == models.WalletStatusPending {
				pendingWallet = wallet
				break
			}
		}
		require.NotNil(t, pendingWallet)

		token := suite.getAuthToken(t)

		// Activate wallet via API
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/wallets/%s/activate", pendingWallet.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify wallet is now active
		activatedWallet, err := suite.walletService.GetWalletByID(pendingWallet.ID)
		require.NoError(t, err)
		assert.Equal(t, models.WalletStatusActive, activatedWallet.Status)
	})

	t.Run("Wallet Whitelisting Workflow", func(t *testing.T) {
		// Get an active wallet
		wallets, err := suite.walletService.GetWalletsForEnterprise(suite.testEnterprise.ID)
		require.NoError(t, err)

		var activeWallet *models.WalletResponse
		for _, wallet := range wallets {
			if wallet.Status == models.WalletStatusActive {
				activeWallet = wallet
				break
			}
		}
		require.NotNil(t, activeWallet)

		token := suite.getAuthToken(t)

		// Whitelist wallet via API
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/wallets/%s/whitelist", activeWallet.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify wallet is now whitelisted
		whitelistedWallet, err := suite.walletService.GetWalletByID(activeWallet.ID)
		require.NoError(t, err)
		assert.True(t, whitelistedWallet.IsWhitelisted)
	})

	t.Run("Wallet Health Monitoring", func(t *testing.T) {
		token := suite.getAuthToken(t)

		// Get wallet health status
		req := httptest.NewRequest("GET", "/api/wallets/health-status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		healthStatus := response["health_status"].(map[string]interface{})
		assert.Greater(t, int(healthStatus["total_wallets"].(float64)), 0)
		assert.GreaterOrEqual(t, int(healthStatus["active_wallets"].(float64)), 0)
		assert.GreaterOrEqual(t, int(healthStatus["whitelisted_wallets"].(float64)), 0)
	})
}

func TestWalletProvisioningEdgeCases(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	t.Run("Prevent Duplicate Active Wallets", func(t *testing.T) {
		// First, activate the existing testnet wallet
		wallets, err := suite.walletService.GetWalletsForEnterprise(suite.testEnterprise.ID)
		require.NoError(t, err)
		require.Len(t, wallets, 1)

		// Activate the wallet
		err = suite.walletService.ActivateWallet(wallets[0].ID)
		require.NoError(t, err)

		// Now try to create another testnet wallet for the same enterprise - this should fail
		_, err = suite.walletService.CreateWalletForEnterprise(suite.testEnterprise.ID, "testnet")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already has an active wallet")
	})

	t.Run("Wallet Address Validation", func(t *testing.T) {
		// Get a wallet and validate its address
		wallets, err := suite.walletService.GetWalletsForEnterprise(suite.testEnterprise.ID)
		require.NoError(t, err)
		require.Len(t, wallets, 1)

		isValid := suite.walletService.ValidateWalletAddress(wallets[0].Address)
		assert.True(t, isValid)

		// Test invalid address
		isValid = suite.walletService.ValidateWalletAddress("invalid-address")
		assert.False(t, isValid)
	})

	t.Run("Cannot Whitelist Inactive Wallet", func(t *testing.T) {
		// Create a new wallet and try to whitelist without activation
		newWallet, err := suite.walletService.CreateWalletForEnterprise(suite.testEnterprise.ID, "mainnet")
		require.NoError(t, err)

		err = suite.walletService.WhitelistWallet(newWallet.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot whitelist inactive wallet")
	})
}

func getTestPostgresURL() string {
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	
	return fmt.Sprintf("postgres://user:password@%s:5432/smart_payment?sslmode=disable", dbHost)
}

func cleanupTestData(t *testing.T, db *sql.DB) {
	queries := []string{
		"DELETE FROM wallets WHERE enterprise_id IN (SELECT id FROM enterprises WHERE legal_name = 'Test Enterprise Ltd')",
		"DELETE FROM authorized_representatives WHERE enterprise_id IN (SELECT id FROM enterprises WHERE legal_name = 'Test Enterprise Ltd')",
		"DELETE FROM enterprises WHERE legal_name = 'Test Enterprise Ltd'",
		"DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email = 'test@example.com')",
		"DELETE FROM users WHERE email = 'test@example.com'",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if t != nil && err != nil {
			t.Logf("Warning: cleanup query failed: %v", err)
		}
	}
}
