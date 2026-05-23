package config

import "testing"

func TestValidateRejectsWildcardCORSInProduction(t *testing.T) {
	cfg := validTestConfig()
	cfg.AppEnv = "production"
	cfg.CORSAllowedOrigins = []string{"*"}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate error = nil, want wildcard CORS rejection")
	}
}

func TestValidateRejectsHTTPOriginInProduction(t *testing.T) {
	cfg := validTestConfig()
	cfg.AppEnv = "production"
	cfg.CORSAllowedOrigins = []string{"http://app.example.com"}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate error = nil, want non-HTTPS production CORS rejection")
	}
}

func TestValidateAllowsExplicitHTTPSOriginInProduction(t *testing.T) {
	cfg := validTestConfig()
	cfg.AppEnv = "production"
	cfg.CORSAllowedOrigins = []string{"https://app.example.com"}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate error = %v, want nil", err)
	}
}

func validTestConfig() Config {
	return Config{
		AppEnv:             "development",
		HTTPPort:           "8080",
		DatabaseURL:        "postgres://postgres:postgres@localhost:5432/umkm_os?sslmode=disable",
		JWTSecret:          "test-secret",
		AccessTokenTTL:     15,
		RefreshTokenTTL:    30,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
		StorageDriver:      "local",
		StorageLocalDir:    "uploads",
		StoragePublicURL:   "http://localhost:8080/uploads",
		UploadMaxBytes:     5 * 1024 * 1024,
		WorkerPollInterval: 5,
		WorkerBatchSize:    10,
		WorkerMaxAttempts:  5,
		WorkerRetryDelay:   30,
	}
}
