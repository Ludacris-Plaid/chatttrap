package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseServiceRoleKey string

	GensmsAPIKey      string
	FeatherlessAPIKey string
	FeatherlessModel  string
	FeatherlessPrompt string

	CNAMAPIKey        string

	NowPaymentsAPIKey     string
	NowPaymentsIPNSecret  string
	NowPaymentsCallbackURL string

	TokenRate         float64
	VIPPrice          float64
	VIPDurationDays   int

	OTPGrabCostUSD    float64
	OTPGrabTokens     int
	OTPGrabCallCostUSD float64
	OTPGrabCallTokens  int

	AllowedOrigins    []string
	AdminEmail        string

		// SIP Telephony
		SIPHost        string
		SIPPort        string
		SIPUsername    string
		SIPPassword    string
		SIPTransport   string
		SIPDisplayName string
		SIPCallerID    string

		// Logging
		LogLevel string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:         getEnv("JWT_SECRET", "your_jwt_secret_key_here_must_be_at_least_32_chars"),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:   getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		GensmsAPIKey:      getEnv("GENSMS_API_KEY", ""),
		FeatherlessAPIKey: getEnv("FEATHERLESS_API_KEY", ""),
		FeatherlessModel:  getEnv("FEATHERLESS_MODEL", "captain-eris-violet-12b"),
		FeatherlessPrompt: getEnv("FEATHERLESS_PROMPT", ""),
		CNAMAPIKey:        getEnv("CNAM_API_KEY", ""),
		NowPaymentsAPIKey:     getEnv("NOWPAYMENTS_API_KEY", getEnv("NOW_PAYMENTS_API_KEY", "")),
		NowPaymentsIPNSecret:  getEnv("NOWPAYMENTS_IPN_SECRET", getEnv("NOW_PAYMENTS_WEBHOOK_SECRET", "")),
		NowPaymentsCallbackURL: getEnv("NOWPAYMENTS_CALLBACK_URL", "http://localhost:8080/api/webhooks/nowpayments"),
		TokenRate:         getEnvFloat("TOKEN_RATE", 0.50),
		VIPPrice:          getEnvFloat("VIP_PRICE", 250.0),
		VIPDurationDays:   getEnvInt("VIP_DURATION_DAYS", 7),
		OTPGrabCostUSD:    getEnvFloat("OTP_GRAB_COST_USD", 5.0),
		OTPGrabTokens:     getEnvInt("OTP_GRAB_TOKENS", 10),
		OTPGrabCallCostUSD: getEnvFloat("OTP_GRAB_CALL_COST_USD", 2.50),
		OTPGrabCallTokens:  getEnvInt("OTP_GRAB_CALL_TOKENS", 5),
		AllowedOrigins:    strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		AdminEmail:        getEnv("ADMIN_EMAIL", "admin@hushcircuits.io"),
		SIPHost:        getEnv("SIP_HOST", "sip.sipup.org"),
		SIPPort:        getEnv("SIP_PORT", "5060"),
		SIPUsername:    getEnv("SIP_USERNAME", ""),
		SIPPassword:    getEnv("SIP_PASSWORD", ""),
		SIPTransport:   getEnv("SIP_TRANSPORT", "udp"),
		SIPDisplayName: getEnv("SIP_DISPLAY_NAME", "HushCircuits"),
		SIPCallerID:    getEnv("SIP_CALLER_ID", ""),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
