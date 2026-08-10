// internal/shared/config/config.go

package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Email       EmailConfig
	Casbin      CasbinConfig
	MPesa       MPesaConfig
	Supabase	SupabaseConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type SupabaseConfig struct {
	URL              string
	SecretKey        string
	PublishableKey   string  // Add this
	BucketEvent      string
	BucketBusiness   string
	BucketProfile    string
	BucketCertificate string
	BucketRecording  string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret            string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

type EmailConfig struct {
	APIKey string
	From   string
}

type CasbinConfig struct {
	ModelPath        string
	AutoLoad         bool
	AutoLoadInterval time.Duration
}

type MPesaConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	Passkey        string
	Shortcode      string
	Environment    string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "54582"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "nuruvent"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "localhost:6379"),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "change-this-in-production"),
			AccessExpiration:  getEnvDuration("JWT_ACCESS_EXPIRATION", 24*time.Hour),
			RefreshExpiration: getEnvDuration("JWT_REFRESH_EXPIRATION", 168*time.Hour),
		},
		Email: EmailConfig{
			APIKey: getEnv("RESEND_API_KEY", ""),
			From:   getEnv("EMAIL_FROM", "noreply@nuruvent.com"),
		},
		Casbin: CasbinConfig{
			ModelPath:        getEnv("CASBIN_MODEL", "configs/casbin/model.conf"),
			AutoLoad:         getEnvBool("CASBIN_AUTO_LOAD", true),
			AutoLoadInterval: getEnvDuration("CASBIN_AUTO_LOAD_INTERVAL", 10*time.Second),
		},
		MPesa: MPesaConfig{
			ConsumerKey:    getEnv("MPESA_CONSUMER_KEY", ""),
			ConsumerSecret: getEnv("MPESA_CONSUMER_SECRET", ""),
			Passkey:        getEnv("MPESA_PASSKEY", ""),
			Shortcode:      getEnv("MPESA_SHORTCODE", "174379"),
			Environment:    getEnv("MPESA_ENVIRONMENT", "sandbox"),
		},
		Supabase: SupabaseConfig{
			URL:              getEnv("SUPABASE_URL", ""),
			SecretKey:        getEnv("SUPABASE_SECRET_KEY", ""),
			PublishableKey:   getEnv("SUPABASE_PUBLISHABLE_KEY", ""), 
			BucketEvent:      getEnv("SUPABASE_BUCKET_EVENTS", "events"),
			BucketBusiness:   getEnv("SUPABASE_BUCKET_BUSINESSES", "businesses"),
			BucketProfile:    getEnv("SUPABASE_BUCKET_PROFILES", "profiles"),
			BucketCertificate: getEnv("SUPABASE_BUCKET_CERTIFICATES", "certificates"),
			BucketRecording:  getEnv("SUPABASE_BUCKET_RECORDINGS", "recordings"),
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
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}