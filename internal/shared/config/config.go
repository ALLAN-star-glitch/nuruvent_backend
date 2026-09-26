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
	Paystack    PaystackConfig
	Zoom        ZoomConfig
	Video       VideoConfig
	App         AppConfig
	Supabase    SupabaseConfig
	OpenAI      OpenAIConfig
	Gemini      GeminiConfig
	Groq        GroqConfig
	OpenRouter  OpenRouterConfig
	NuruOnboardingNoticeEmails NuruventOnboardingNoticeEmails
}

// ============================================================
// PAYSTACK
// ============================================================

// PaystackConfig holds credentials for the Paystack payment gateway.
type PaystackConfig struct {
	SecretKey string
	PublicKey string
	BaseURL   string
	Enabled   bool
}

func (c PaystackConfig) IsConfigured() bool {
	return c.Enabled && c.SecretKey != "" && c.PublicKey != ""
}

// ============================================================
// ZOOM (webhook-only)
// ============================================================

// ZoomConfig holds credentials for the Zoom webhook integration.
//
// This config is used by the attendance module's Zoom webhook
// provider to verify signatures on incoming participant events.
//
// The OAuth credentials used for connecting host accounts live in
// VideoConfig.ZoomOAuth, because they belong to the video module.
type ZoomConfig struct {
	// SecretToken is the app's Secret Token from the Zoom
	// Marketplace's Feature → Event Subscriptions page. It signs
	// every webhook delivery (HMAC-SHA256).
	SecretToken string

	// Enabled turns webhook processing on or off. When false, the
	// attendance module will not register a Zoom provider.
	Enabled bool
}

// IsConfigured reports whether the minimum required credentials are
// present for the Zoom webhook integration to operate.
func (c ZoomConfig) IsConfigured() bool {
	return c.Enabled && c.SecretToken != ""
}

// ============================================================
// VIDEO (OAuth + encryption)
// ============================================================

// VideoConfig holds everything the video module needs to operate:
// per-platform OAuth credentials and the encryption key used to
// protect tokens at rest.
type VideoConfig struct {
	// EncryptionKey is a base64- or hex-encoded 32-byte key used to
	// encrypt OAuth tokens before storing them.
	//
	// Required. Without it, the video module cannot start.
	EncryptionKey string

	// ZoomOAuth carries the Zoom General App's OAuth credentials.
	// Used when hosts connect their own Zoom accounts.
	ZoomOAuth VideoOAuthPlatformConfig

	// GoogleMeetOAuth is reserved for the eventual Google Meet
	// integration. Leaving it unconfigured keeps the platform
	// disabled.
	GoogleMeetOAuth VideoOAuthPlatformConfig
}

// VideoOAuthPlatformConfig is the set of OAuth credentials for one
// video platform.
type VideoOAuthPlatformConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Enabled      bool
}

// IsConfigured reports whether the platform has enough credentials to
// operate.
func (c VideoOAuthPlatformConfig) IsConfigured() bool {
	return c.Enabled && c.ClientID != "" && c.ClientSecret != "" && c.RedirectURI != ""
}

// IsConfigured reports whether the video module has the minimum
// required credentials. Only the encryption key is strictly required;
// individual platforms enable themselves.
func (c VideoConfig) IsConfigured() bool {
	return c.EncryptionKey != ""
}

// ============================================================
// APP
// ============================================================

// AppConfig holds application-level settings that aren't tied to a
// specific subsystem.
type AppConfig struct {
	// PublicURL is the base URL of the public-facing app. Used to
	// build join links, confirmation links, etc.
	PublicURL string
}

// ============================================================
// OTHER SUBSYSTEMS
// ============================================================

type GroqConfig struct {
	APIKey string
	Model  string
}

type OpenRouterConfig struct {
	APIKey string
	Model  string
}

type NuruventOnboardingNoticeEmails struct {
	AdminEmail          string
	MarketingEmail      string
	OnboardingTeamEmail string
	CeoEmail            string
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

type OpenAIConfig struct {
	APIKey string
	Model  string
}

type GeminiConfig struct {
	APIKey string
	Model  string
}

// ============================================================
// LOAD
// ============================================================

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	} else {
		log.Println("✅ .env file loaded successfully")
	}

	redisURL := getEnv("REDIS_URL", "redis://nuruvent-redis:6379")

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Redis: RedisConfig{
			URL: redisURL,
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
		Paystack: PaystackConfig{
			SecretKey: getEnv("PAYSTACK_SECRET_KEY", ""),
			PublicKey: getEnv("PAYSTACK_PUBLIC_KEY", ""),
			BaseURL:   getEnv("PAYSTACK_BASE_URL", "https://api.paystack.co"),
			Enabled:   getEnvBool("PAYSTACK_ENABLED", false),
		},
		Zoom: ZoomConfig{
			SecretToken: getEnv("ZOOM_SECRET_TOKEN", ""),
			Enabled:     getEnvBool("ZOOM_ENABLED", false),
		},
		Video: VideoConfig{
			EncryptionKey: getEnv("VIDEO_TOKEN_ENCRYPTION_KEY", ""),
			ZoomOAuth: VideoOAuthPlatformConfig{
				ClientID:     getEnv("ZOOM_OAUTH_CLIENT_ID", ""),
				ClientSecret: getEnv("ZOOM_OAUTH_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("ZOOM_OAUTH_REDIRECT_URI", ""),
				Enabled:      getEnvBool("ZOOM_OAUTH_ENABLED", true),
			},
			GoogleMeetOAuth: VideoOAuthPlatformConfig{
				ClientID:     getEnv("GOOGLE_MEET_OAUTH_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_MEET_OAUTH_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("GOOGLE_MEET_OAUTH_REDIRECT_URI", ""),
				Enabled:      getEnvBool("GOOGLE_MEET_OAUTH_ENABLED", false),
			},
		},
		App: AppConfig{
			PublicURL: getEnv("APP_PUBLIC_URL", "http://localhost:3000"),
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
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
			Model:  getEnv("OPENAI_MODEL", "gpt-3.5-turbo"),
		},
		Gemini: GeminiConfig{
			APIKey: getEnv("GEMINI_API_KEY", ""),
			Model:  getEnv("GEMINI_MODEL", "gemini-3.6-flash"),
		},
		Groq: GroqConfig{
			APIKey: getEnv("GROQ_API_KEY", ""),
			Model:  getEnv("GROQ_MODEL", "llama-3.3-70b-versatile"),
		},
		OpenRouter: OpenRouterConfig{
			APIKey: getEnv("OPENROUTER_API_KEY", ""),
			Model:  getEnv("OPENROUTER_MODEL", "meta-llama/llama-3.1-8b-instruct:free"),
		},
		NuruOnboardingNoticeEmails: NuruventOnboardingNoticeEmails{
			AdminEmail:          getEnv("NURUVENT_ADMIN_EMAIL", "allaneditor67@gmail.com"),
			OnboardingTeamEmail: getEnv("NURUVENT_ONBOARDING_TEAM_EMAIL", "allanmathenge22@gmail.com"),
			MarketingEmail:      getEnv("NURUVENT_MARKETING_EMAIL", "allanmathenge82@gmail.com"),
			CeoEmail:            getEnv("NURUVENT_CEO_EMAIL", "allanmathenge67@gmail.com"),
		},
	}

	// Database: prefer DATABASE_URL, fall back to individual fields.
	dbURL := getEnv("DATABASE_URL", "")
	if dbURL != "" {
		cfg.Database.URL = dbURL
		log.Println("✅ Using DATABASE_URL for database connection")
	} else {
		cfg.Database.Host = getEnv("DB_HOST", "localhost")
		cfg.Database.Port = getEnv("DB_PORT", "5432")
		cfg.Database.User = getEnv("DB_USER", "postgres")
		cfg.Database.Password = getEnv("DB_PASSWORD", "")
		cfg.Database.Name = getEnv("DB_NAME", "nuruvent")
		cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")
		log.Println("✅ Using individual DB fields for database connection")
	}

	return cfg
}

// ============================================================
// HELPERS
// ============================================================

func (c *Config) GetDSN() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
	return "postgres://" + c.Database.User + ":" + c.Database.Password +
		"@" + c.Database.Host + ":" + c.Database.Port +
		"/" + c.Database.Name + "?sslmode=" + c.Database.SSLMode
}

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