//go:build oci_integration
// +build oci_integration

package ora

import (
	"context"
	"database/sql/driver"
	"errors"
	"os"
	"sync/atomic"
	"testing"
)

// This native-client test does not connect to a database.
func TestOCIEnvironmentHandles(t *testing.T) {
	if os.Getenv("ORA_TEST_NATIVE_HANDLES") != "1" {
		t.Skip("set ORA_TEST_NATIVE_HANDLES=1 with Instant Client installed")
	}
	before := atomic.LoadInt64(&handleRefCount)
	for i := 0; i < 20; i++ {
		conn, err := newConnection()
		if err != nil {
			t.Fatal(err)
		}
		conn.free()
	}
	if got := atomic.LoadInt64(&handleRefCount); got != before {
		t.Fatalf("native handles leaked: before=%d after=%d", before, got)
	}
}

func TestOCIHandleFreeFailurePreservesCount(t *testing.T) {
	if os.Getenv("ORA_TEST_NATIVE_HANDLES") != "1" {
		t.Skip("set ORA_TEST_NATIVE_HANDLES=1 with Instant Client installed")
	}
	before := atomic.LoadInt64(&handleRefCount)
	handle := &ociHandle{typ: OCI_HTYPE_ENV}
	handle.free() // A null handle cannot be freed.
	if got := atomic.LoadInt64(&handleRefCount); got != before {
		t.Fatalf("failed free changed handle count: before=%d after=%d", before, got)
	}
}

func TestOCIBeginFailureReleasesHandles(t *testing.T) {
	if os.Getenv("ORA_TEST_NATIVE_HANDLES") != "1" {
		t.Skip("set ORA_TEST_NATIVE_HANDLES=1 with Instant Client installed")
	}
	conn, err := newConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.free()
	before := atomic.LoadInt64(&handleRefCount)
	// No logon: attaching a transaction to the null service context must fail cleanly.
	if _, err := conn.BeginTx(context.Background(), driver.TxOptions{ReadOnly: true}); err == nil {
		t.Fatal("begin succeeded without a service context")
	}
	if conn.tx != nil || conn.trans != nil {
		t.Fatal("failed begin retained transaction state")
	}
	if got := atomic.LoadInt64(&handleRefCount); got != before {
		t.Fatalf("failed begin leaked handles: before=%d after=%d", before, got)
	}
}

func TestOCIPingFailureInvalidatesConnection(t *testing.T) {
	if os.Getenv("ORA_TEST_NATIVE_HANDLES") != "1" {
		t.Skip("set ORA_TEST_NATIVE_HANDLES=1 with Instant Client installed")
	}
	conn, err := newConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.free()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := conn.Ping(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled ping: %v", err)
	}
	if conn.bad {
		t.Fatal("pre-canceled ping invalidated the connection")
	}

	// No logon: the native ping fails with an invalid service handle.
	if err := conn.Ping(context.Background()); !errors.Is(err, driver.ErrBadConn) {
		t.Fatalf("native ping failure: got %v, want ErrBadConn", err)
	}
	if !conn.bad {
		t.Fatal("native ping failure did not invalidate the connection")
	}
	if err := conn.ResetSession(context.Background()); !errors.Is(err, driver.ErrBadConn) {
		t.Fatalf("reset after failed ping: %v", err)
	}
}
