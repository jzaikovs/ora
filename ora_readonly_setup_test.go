//go:build oracle_integration && !oracle_write_integration
// +build oracle_integration,!oracle_write_integration

package ora

// The read-only suite never creates or drops database objects.
func setupOracleIntegration() error { return nil }

func teardownOracleIntegration() error { return nil }
