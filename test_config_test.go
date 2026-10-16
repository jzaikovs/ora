package ora

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type legacyTestConfig struct {
	user     string
	password string
	host     string
}

// Only the legacy TestMain calls this; normal driver use never reads .env.
func loadLegacyTestConfig(path string) (legacyTestConfig, error) {
	values, err := godotenv.Read(path)
	if err != nil && !os.IsNotExist(err) {
		// Parser errors may contain source text, including credentials.
		return legacyTestConfig{}, fmt.Errorf("ora: cannot read or parse test environment file %q", path)
	}
	value := func(key string) string {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return values[key]
	}
	config := legacyTestConfig{
		user:     value("ORA_TEST_USER"),
		password: value("ORA_TEST_PASSWORD"),
		host:     value("ORA_TEST_HOST"),
	}
	missing := []string{}
	for _, field := range []struct{ key, value string }{
		{key: "ORA_TEST_USER", value: config.user},
		{key: "ORA_TEST_PASSWORD", value: config.password},
		{key: "ORA_TEST_HOST", value: config.host},
	} {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.key)
		}
	}
	if len(missing) != 0 {
		return legacyTestConfig{}, fmt.Errorf(
			"ora: required test configuration missing: %s; "+
				"set environment variables or fill in .env using .env.example",
			strings.Join(missing, ", "),
		)
	}
	return config, nil
}
