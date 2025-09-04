package config

import (
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	Redis    RedisConfig
	Database DatabaseConfig
	JWT      JWTConfig
	XRPL     XRPLConfig
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	PostgresURL string
	MongoURL    string
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  string
	RefreshTokenDuration string
}

// XRPLConfig represents XRPL configuration
type XRPLConfig struct {
	NetworkURL   string
	WebSocketURL string
	TestNet      bool
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Database: DatabaseConfig{
			PostgresURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/smartcheques?sslmode=disable"),
			MongoURL:    getEnv("MONGO_URL", "mongodb://localhost:27017"),
		},
		JWT: JWTConfig{
			SecretKey:            getEnv("JWT_SECRET_KEY", "your-secret-key"),
			AccessTokenDuration:  getEnv("JWT_ACCESS_TOKEN_DURATION", "15m"),
			RefreshTokenDuration: getEnv("JWT_REFRESH_TOKEN_DURATION", "24h"),
		},
		XRPL: XRPLConfig{
			NetworkURL:   getEnv("XRPL_NETWORK_URL", "https://s.altnet.rippletest.net:51234"),
			WebSocketURL: getEnv("XRPL_WEBSOCKET_URL", "wss://s.altnet.rippletest.net:51233"),
			TestNet:      getEnvAsBool("XRPL_TESTNET", true),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
