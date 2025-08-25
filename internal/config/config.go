package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	XRPL     XRPLConfig
	Server   ServerConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type DatabaseConfig struct {
	PostgresURL string
	MongoURL    string
}

type XRPLConfig struct {
	NetworkURL string
	TestNet    bool
}

type ServerConfig struct {
	Port string
	Env  string
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  string
	RefreshTokenDuration string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			PostgresURL: getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/smart_payment?sslmode=disable"),
			MongoURL:    getEnv("MONGO_URL", "mongodb://localhost:27017"),
		},
		XRPL: XRPLConfig{
			NetworkURL: getEnv("XRPL_NETWORK_URL", "wss://s.altnet.rippletest.net:51233"),
			TestNet:    getEnvBool("XRPL_TESTNET", true),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			SecretKey:            getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			AccessTokenDuration:  getEnv("JWT_ACCESS_TOKEN_DURATION", "15m"),
			RefreshTokenDuration: getEnv("JWT_REFRESH_TOKEN_DURATION", "168h"), // 7 days
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
