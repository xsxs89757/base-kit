package config

import (
	"strings"
	"testing"
)

func withConfig(t *testing.T, mode, secret string) {
	t.Helper()
	saved := C
	t.Cleanup(func() { C = saved })
	C.Server.Mode = mode
	C.JWT.Secret = secret
}

func TestIsProduction(t *testing.T) {
	withConfig(t, " Production ", "")
	if !IsProduction() {
		t.Fatal("expected ' Production ' to be treated as production")
	}
	C.Server.Mode = "development"
	if IsProduction() {
		t.Fatal("development must not be production")
	}
	C.Server.Mode = ""
	if IsProduction() {
		t.Fatal("empty mode must not be production")
	}
}

func TestValidateProductionSkipsDevelopment(t *testing.T) {
	withConfig(t, "development", "x")
	if err := ValidateProduction(); err != nil {
		t.Fatalf("development mode must not validate secret, got %v", err)
	}
}

func TestValidateProductionRejectsPlaceholderAndShortSecrets(t *testing.T) {
	withConfig(t, "production", "")

	for _, placeholder := range placeholderJWTSecrets {
		C.JWT.Secret = strings.ToUpper(placeholder)
		if err := ValidateProduction(); err == nil {
			t.Fatalf("placeholder %q must be rejected", placeholder)
		}
	}

	C.JWT.Secret = strings.Repeat("a", minJWTSecretLen-1)
	if err := ValidateProduction(); err == nil {
		t.Fatalf("%d-char secret must be rejected", minJWTSecretLen-1)
	}

	C.JWT.Secret = "kF3n9xQz2vLm8pRt4wYb6hJc1sDg7aEu5iOe0rNq"
	if err := ValidateProduction(); err != nil {
		t.Fatalf("strong secret must pass, got %v", err)
	}
}
