package app

import (
	"os"
	"strings"
	"testing"
)

func TestValidateStartupConfigRequiresAllCriticalEnv(t *testing.T) {
	keys := []string{
		"SECRET_KEY",
		"CLIENT_ID",
		"SYSTEM",
		"MONGO_HOST",
		"MONGO_POS_DB_NAME",
		"REDIS_HOST",
	}

	restore := snapshotEnv(keys)
	defer restore()

	for _, key := range keys {
		t.Setenv(key, "")
	}

	err := validateStartupConfig()
	if err == nil {
		t.Fatal("expected missing env error")
	}
	if !strings.Contains(err.Error(), "SECRET_KEY") {
		t.Fatalf("expected error to mention first missing key, got %v", err)
	}
}

func TestValidateStartupConfigSucceedsWhenAllCriticalEnvPresent(t *testing.T) {
	keys := []string{
		"SECRET_KEY",
		"CLIENT_ID",
		"SYSTEM",
		"MONGO_HOST",
		"MONGO_POS_DB_NAME",
		"REDIS_HOST",
	}

	restore := snapshotEnv(keys)
	defer restore()

	t.Setenv("SECRET_KEY", "secret")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("SYSTEM", "pos")
	t.Setenv("MONGO_HOST", "mongodb://localhost:27017")
	t.Setenv("MONGO_POS_DB_NAME", "pos")
	t.Setenv("REDIS_HOST", "redis://localhost:6379/0")

	if err := validateStartupConfig(); err != nil {
		t.Fatalf("expected config validation to pass, got %v", err)
	}
}

func snapshotEnv(keys []string) func() {
	values := make(map[string]*string, len(keys))
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			copied := value
			values[key] = &copied
			continue
		}
		values[key] = nil
	}

	return func() {
		for _, key := range keys {
			if values[key] == nil {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, *values[key])
		}
	}
}
