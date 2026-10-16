package ora

import (
	"context"
	"database/sql/driver"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestContextAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := interruptible(ctx, func() error { called = true; return nil }, func() { called = true }, func() { called = true })
	if !errors.Is(err, context.Canceled) || called {
		t.Fatalf("err=%v called=%t", err, called)
	}
	conn := &Conn{}
	if err := conn.Ping(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled ping: %v", err)
	}
	if conn.bad {
		t.Fatal("pre-canceled ping invalidated the connection")
	}
	if _, err := conn.BeginTx(ctx, driver.TxOptions{ReadOnly: true}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := conn.PrepareContext(ctx, "select 1 from dual"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := (&ConnStd{conn}).QueryContext(ctx, "select 1 from dual", nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestContextCancellationJoinsBreaker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered, release := make(chan struct{}), make(chan struct{})
	returned := make(chan error, 1)
	invalid := false
	go func() {
		returned <- interruptible(ctx, func() error { cancel(); <-entered; return nil }, func() { close(entered); <-release }, func() { invalid = true })
	}()
	<-entered
	select {
	case err := <-returned:
		t.Fatalf("returned before breaker finished: %v", err)
	default:
	}
	close(release)
	select {
	case err := <-returned:
		if !errors.Is(err, context.Canceled) || !invalid {
			t.Fatalf("err=%v invalid=%t", err, invalid)
		}
	case <-time.After(time.Second):
		t.Fatal("breaker did not join")
	}
}

func TestContextCancellationBetweenNativeCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first, second := make(chan struct{}), make(chan struct{})
	var calls int32
	err := interruptible(ctx, func() error {
		cancel()
		<-first // First break happens while no simulated native call is running.
		<-second
		return nil
	}, func() {
		switch atomic.AddInt32(&calls, 1) {
		case 1:
			close(first)
		case 2:
			close(second)
		}
	}, func() {})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestContextCanceledAfterSuccessfulWorkInvalidates(t *testing.T) {
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		invalid := false
		err := interruptible(ctx, func() error { cancel(); return nil }, func() {}, func() { invalid = true })
		if !errors.Is(err, context.Canceled) || !invalid {
			t.Fatalf("err=%v invalid=%t", err, invalid)
		}
	}
}

func TestContextCompletionStopsWatcher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls int32
	want := errors.New("work error")
	err := interruptible(ctx, func() error { return want }, func() { atomic.AddInt32(&calls, 1) }, func() { t.Error("invalidated completed connection") })
	cancel()
	if !errors.Is(err, want) || atomic.LoadInt32(&calls) != 0 {
		t.Fatal(err, calls)
	}
}

func TestInvalidConnectionIsNotReused(t *testing.T) {
	conn := &Conn{opened: true, bad: true}
	if conn.IsValid() {
		t.Fatal("interrupted connection is valid")
	}
	if err := conn.ResetSession(context.Background()); !errors.Is(err, driver.ErrBadConn) {
		t.Fatal(err)
	}
}

func TestReadOnlyOptionsRejectUnsupportedIsolation(t *testing.T) {
	if _, err := (&Conn{}).BeginTx(context.Background(), driver.TxOptions{Isolation: 99, ReadOnly: true}); err == nil {
		t.Fatal("unsupported isolation accepted")
	}
}
