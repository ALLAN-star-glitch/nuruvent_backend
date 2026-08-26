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
	Supabase    SupabaseConfig
	NuruOnboardingNoticeEmails	NuruventOnboardingNoticeEmails

}


type NuruventOnboardingNoticeEmails struct{
	AdminEmail string
	MarketingEmail string
	OnboardingTeamEmail string
	CeoEmail string
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
	URL      string
}

type SupabaseConfig struct {
	URL               string
	SecretKey         string
	PublishableKey    string
	BucketEvent       string
	BucketBusiness    string
	BucketProfile     string
	BucketCertificate string
	BucketRecording   string
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
	} else {
		log.Println("✅ .env file loaded successfully")
	}

	// In Load() inside config.go:
	redisURL := getEnv("REDIS_URL", "redis://nuruvent-redis:6379")

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Redis: RedisConfig{
			URL: redisURL, // ✅ Use the variable
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "change-this-in-production"),
			AccessExpiration:  getEnvDuration("JWT_ACCESS_EXPIRATION", 24*time.Hour),
			RefreshExpiration: getEnvDuration("JWT_REFRESH_EXPIRATION", 168*time.Hour),
		},
		Email: EmailConfig{
			APIKey: getEnv("EMAIL_API_KEY", ""),
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
			URL:               getEnv("SUPABASE_URL", ""),
			SecretKey:         getEnv("SUPABASE_SECRET_KEY", ""),
			PublishableKey:    getEnv("SUPABASE_PUBLISHABLE_KEY", ""),
			BucketEvent:       getEnv("SUPABASE_BUCKET_EVENTS", "events"),
			BucketBusiness:    getEnv("SUPABASE_BUCKET_BUSINESSES", "businesses"),
			BucketProfile:     getEnv("SUPABASE_BUCKET_PROFILES", "profiles"),
			BucketCertificate: getEnv("SUPABASE_BUCKET_CERTIFICATES", "certificates"),
			BucketRecording:   getEnv("SUPABASE_BUCKET_RECORDINGS", "recordings"),
		},
		NuruOnboardingNoticeEmails: NuruventOnboardingNoticeEmails{
			AdminEmail: getEnv("NURUVENT_ADMIN_EMAIL", "allaneditor67@gmail.com"),
			OnboardingTeamEmail: getEnv("NURUVENT_ONBOARDING_TEAM_EMAIL", "allanmathenge22@gmail.com"),
			MarketingEmail: getEnv("NURUVENT_MARKETING_EMAIL", "allanmathenge82@gmail.com"),
			CeoEmail: getEnv("NURUVENT_CEO_EMAIL", "allanmathenge67@gmail.com"),
		},
	}

	// Handle Database: Support both DATABASE_URL and individual fields
	dbURL := getEnv("DATABASE_URL", "")
	if dbURL != "" {
		cfg.Database.URL = dbURL
		log.Println("✅ Using DATABASE_URL for database connection")
	} else {
		cfg.Database.Host = getEnv("DB_HOST", "localhost")
		cfg.Database.Port = getEnv("DB_PORT", "54582")
		cfg.Database.User = getEnv("DB_USER", "postgres")
		cfg.Database.Password = getEnv("DB_PASSWORD", "")
		cfg.Database.Name = getEnv("DB_NAME", "nuruvent")
		cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")
		log.Println("✅ Using individual DB fields for database connection")
	}

	return cfg
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
	return "postgres://" + c.Database.User + ":" + c.Database.Password +
		"@" + c.Database.Host + ":" + c.Database.Port +
		"/" + c.Database.Name + "?sslmode=" + c.Database.SSLMode
}

// GetRedisURL returns the Redis connection string
func (c *Config) GetRedisURL() string {
	return c.Redis.URL
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