//go:build oracle_integration || oracle_write_integration
// +build oracle_integration oracle_write_integration

package ora

import (
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
)

var (
	dbUser              string
	dbPassword          string
	dbHost              string
	testDBConnectString string
	db                  *sql.DB
)

func TestMain(m *testing.M) {
	os.Exit(runOracleIntegration(m))
}

func runOracleIntegration(m *testing.M) int {
	config, err := loadOracleTestConfig(".env")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	dbUser, dbPassword, dbHost = config.user, config.password, config.host
	testDBConnectString = fmt.Sprintf("%s/%s@%s", dbUser, dbPassword, dbHost)
	db, err = sql.Open(DriverName, testDBConnectString)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err = db.Ping(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = db.Close()
		return 1
	}
	if err = setupOracleIntegration(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = db.Close()
		return 1
	}
	code := m.Run()
	if err = teardownOracleIntegration(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	if err = db.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	if atomic.LoadInt64(&handleRefCount) != 0 {
		fmt.Fprintln(os.Stderr, "OCI handles not freed")
		code = 1
	}
	return code
}
