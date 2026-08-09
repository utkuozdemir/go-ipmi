package bmc

import (
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
)

// TestV15EvictExpiredNoDeadlockUnderHeldLock reproduces the eviction
// self-deadlock: a goroutine holding a v1.5 session lock that then triggers
// eviction (as Get Session Challenge -> CreatePending -> EvictExpired does when
// dispatched under a held session lock) must not block on its own session lock.
func TestV15EvictExpiredNoDeadlockUnderHeldLock(t *testing.T) {
	store := NewV15SessionStore(clock.Real)
	us := NewUserStore()
	u, err := us.Add(2, "admin")
	if err != nil {
		t.Fatal(err)
	}

	sess, err := store.CreatePending(V15AuthTypeMD5, u, [16]byte{}, 1)
	if err != nil {
		t.Fatal(err)
	}

	sess.Lock()
	defer sess.Unlock()

	done := make(chan struct{})
	go func() {
		store.EvictExpired()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("DEADLOCK: EvictExpired blocked on a session lock held by the caller")
	}
}

// TestRMCPPlusEvictExpiredNoDeadlockUnderHeldLock is the RMCP+ counterpart: the
// same self-deadlock would exist in the v2.0 session store's eviction path.
func TestRMCPPlusEvictExpiredNoDeadlockUnderHeldLock(t *testing.T) {
	store := NewSessionStore(clock.Real)

	sess, err := store.Allocate(0xABCD1234, 0, 0, 0, PrivilegeLevelAdministrator, 1)
	if err != nil {
		t.Fatal(err)
	}

	sess.Lock()
	defer sess.Unlock()

	done := make(chan struct{})
	go func() {
		store.EvictExpired()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("DEADLOCK: EvictExpired blocked on a session lock held by the caller")
	}
}
