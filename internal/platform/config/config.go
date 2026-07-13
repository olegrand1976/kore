package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr             string
	DatabaseURL          string
	RedisAddr            string
	RedisAuth            string
	RedisDB              int
	RedisKeyPrefix       string
	RedisTLS             bool
	CacheDefaultTTL      time.Duration
	JWTSigningKey        string
	JWTTTL               time.Duration
	JWTRefreshTTL        time.Duration
	MigrateOnBoot        bool
	LogLevel             string
	SMTPHost             string
	SMTPPort             int
	SMTPFrom             string
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePublishableKey string
	BillingTrialDays     int
	DevSeedEnabled       bool
	UploadsDir           string
	PlatformAdminLogins  []string
	AILLMProvider        string
	GeminiAPIKey         string
	GeminiModel          string
	PromptGuardBlock     bool
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:             envOr("HTTP_ADDR", ":8080"),
		DatabaseURL:          envOr("DATABASE_URL", "postgres://kore:kore@localhost:5432/kore?sslmode=disable"),
		RedisAddr:            envOr("REDIS_ADDR", "localhost:6379"),
		RedisAuth:            envOr("REDIS_AUTH", ""),
		RedisDB:              envInt("REDIS_DB", 0),
		RedisKeyPrefix:       envOr("REDIS_KEY_PREFIX", "kore"),
		RedisTLS:             envBool("REDIS_TLS", false),
		CacheDefaultTTL:      envDuration("CACHE_DEFAULT_TTL", 5*time.Minute),
		JWTSigningKey:        envOr("JWT_SIGNING_KEY", "dev-insecure-change-me"),
		JWTTTL:               envDuration("JWT_TTL", 15*time.Minute),
		JWTRefreshTTL:        envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		MigrateOnBoot:        envBool("MIGRATE_ON_BOOT", false),
		LogLevel:             envOr("LOG_LEVEL", "info"),
		SMTPHost:             envOr("SMTP_HOST", "localhost"),
		SMTPPort:             envInt("SMTP_PORT", 1025),
		SMTPFrom:             envOr("SMTP_FROM", "noreply@kore.local"),
		StripeSecretKey:      envOr("STRIPE_SECRET_KEY", "sk_test_mock"),
		StripeWebhookSecret:  envOr("STRIPE_WEBHOOK_SECRET", "whsec_test"),
		StripePublishableKey: envOr("STRIPE_PUBLISHABLE_KEY", "pk_test_mock"),
		BillingTrialDays:     envInt("BILLING_TRIAL_DAYS", 14),
		DevSeedEnabled:       envBool("DEV_SEED_ENABLED", true),
		UploadsDir:           envOr("UPLOADS_DIR", "./uploads"),
		PlatformAdminLogins:  envCSV("PLATFORM_ADMIN_LOGINS", "ADM_admin"),
		AILLMProvider:        envOr("AI_LLM_PROVIDER", "stub"),
		GeminiAPIKey:         envOr("GEMINI_API_KEY", ""),
		GeminiModel:          envOr("GEMINI_MODEL", "gemini-3.5-flash"),
		PromptGuardBlock:     envBool("PROMPT_GUARD_BLOCK", true),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func envCSV(key, fallback string) []string {
	raw := envOr(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
