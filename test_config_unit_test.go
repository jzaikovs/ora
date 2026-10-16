package ora

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOracleTestConfig(t *testing.T) {
	for _, key := range []string{"ORA_TEST_USER", "ORA_TEST_PASSWORD", "ORA_TEST_HOST"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), ".env")
	contents := "ORA_TEST_USER=TEST_USER\nORA_TEST_PASSWORD='p@ss/#:$word'\nORA_TEST_HOST=//test-host:1521/TEST\n"
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	want := oracleTestConfig{
		user:     "TEST_USER",
		password: "p@ss/#:$word",
		host:     "//test-host:1521/TEST",
	}
	got, err := loadOracleTestConfig(path)
	if err != nil || got != want {
		t.Fatalf("file configuration did not preserve values: err=%v", err)
	}

	t.Setenv("ORA_TEST_HOST", "//override:1521/OTHER")
	got, err = loadOracleTestConfig(path)
	want.host = "//override:1521/OTHER"
	if err != nil || got != want {
		t.Fatalf("environment did not override only the specified field: err=%v", err)
	}

	t.Setenv("ORA_TEST_PASSWORD", "")
	if _, err := loadOracleTestConfig(path); err == nil || !strings.Contains(err.Error(), "ORA_TEST_PASSWORD") {
		t.Fatalf("explicit empty environment value must not fall back to file: %v", err)
	}

	t.Setenv("ORA_TEST_USER", "ENV_USER")
	t.Setenv("ORA_TEST_PASSWORD", "environment-secret")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	got, err = loadOracleTestConfig(path)
	want = oracleTestConfig{
		user:     "ENV_USER",
		password: "environment-secret",
		host:     "//override:1521/OTHER",
	}
	if err != nil || got != want {
		t.Fatalf("environment-only configuration failed: err=%v", err)
	}

	for _, key := range []string{"ORA_TEST_USER", "ORA_TEST_PASSWORD", "ORA_TEST_HOST"} {
		t.Setenv(key, "")
	}
	_, err = loadOracleTestConfig(path)
	if err == nil {
		t.Fatal("missing configuration accepted")
	}
	for _, key := range []string{"ORA_TEST_USER", "ORA_TEST_PASSWORD", "ORA_TEST_HOST"} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("missing field not reported: %s", key)
		}
	}
}

func TestOracleTestConfigMalformedFileDoesNotExposeSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("ORA_TEST_PASSWORD='private-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := loadOracleTestConfig(path)
	if err == nil || strings.Contains(err.Error(), "private-secret") {
		t.Fatal("malformed file must fail without exposing credentials")
	}
}
