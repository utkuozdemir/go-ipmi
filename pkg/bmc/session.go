package bmc

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// Session inactivity timeout per IPMI spec:
//   - v1.5 §6.11.13 Session Inactivity Timeout
//   - v2.0 §6.12.15 Session Inactivity Timeout
const DefaultInactivityTimeout = 60 * time.Second

// DefaultSessionEvictInterval is how often the server scans for idle sessions.
// The spec defines the 60-second inactivity limit, not the scan period.
const DefaultSessionEvictInterval = 3 * time.Second

// DefaultInactivityTimeoutTolerance is the LAN inactivity tolerance per
// IPMI v1.5 Table 6-7 (+/- 3 seconds).
const DefaultInactivityTimeoutTolerance = 3 * time.Second

// MaxSessions is the minimum number of concurrent sessions required by the spec.
const MaxSessions = 4

// ErrNoSession is returned when the session ID is not in the store.
var ErrNoSession = errors.New("session not found")

// ErrSessionFull is returned when the store has reached capacity.
var ErrSessionFull = errors.New("no session slots available")

// SessionState tracks which phase of session negotiation has been reached.
type SessionState uint8

const (
	// SessionStatePending means Open Session was received but RAKP is incomplete.
	SessionStatePending SessionState = iota
	// SessionStateActive means RAKP completed and commands may flow.
	SessionStateActive
	// SessionStateClosed means the session was explicitly closed or timed out.
	SessionStateClosed
)

// Session holds all state for one active or pending IPMI session.
//
// Concurrency: the session mutex guards the per-packet and handshake fields that
// change while a packet is dispatched: the inbound/outbound sequence counters,
// the derived keys (SIK/K1/K2), the RAKP nonces (ConsoleRand/BMCRand), Role,
// State, User, PrivilegeLevel, and writes to LastActivity. MaxPrivilege, Channel
// and CreatedAt are set before the session is published into the store and are
// not written again. The store's eviction and count helpers read the fields they
// need (State, CreatedAt, LastActivity) under the session lock, taken with
// TryLock so eviction never blocks on a busy session. The lock order is always
// session-then-store: a caller may take the store lock while holding the session
// lock (eviction does), never the reverse.
type Session struct {
	// mu serializes access to all fields below. See the type doc for the lock
	// order relative to the store lock.
	mu sync.Mutex

	// BMCID is the session ID assigned by the BMC (sent in Open Session Response).
	// It is set once at allocation and never changes afterwards.
	BMCID uint32
	// ConsoleID is the session ID chosen by the remote console.
	ConsoleID uint32

	State SessionState

	// Negotiated algorithms
	AuthAlg      types.AuthAlg
	IntegrityAlg types.IntegrityAlg
	CryptAlg     types.CryptAlg

	// Sequence tracking.
	// InboundSeq is the last accepted sequence number from the console.
	// OutboundSeq is the next sequence number the BMC will use.
	InboundSeq  uint32
	OutboundSeq uint32

	// Session keys derived during RAKP.
	SIK []byte
	K1  []byte
	K2  []byte

	// RAKP exchange state (zeroed once session is active).
	ConsoleRand [16]byte
	BMCRand     [16]byte
	Role        uint8 // whole byte from RAKP1, used in HMAC input

	// User and privilege
	User           *User
	PrivilegeLevel PrivilegeLevel
	MaxPrivilege   PrivilegeLevel

	// Channel this session arrived on.
	Channel uint8

	// Timing
	CreatedAt    time.Time
	LastActivity time.Time
}

// SessionStore is a thread-safe registry of active and pending sessions.
type SessionStore struct {
	mu       sync.Mutex
	sessions map[uint32]*Session
	max      int
	timeout  time.Duration
	clock    clock.Clock
}

// NewSessionStore creates a SessionStore limited to [MaxSessions] concurrent sessions
// with the default inactivity timeout.
func NewSessionStore(clk clock.Clock) *SessionStore {
	return &SessionStore{
		sessions: make(map[uint32]*Session, MaxSessions),
		max:      MaxSessions,
		timeout:  DefaultInactivityTimeout,
		clock:    clk,
	}
}

// Option configures a [SessionStore].
type SessionStoreOption func(*SessionStore)

// WithMaxSessions overrides the default session limit.
func WithMaxSessions(n int) SessionStoreOption {
	return func(s *SessionStore) { s.max = n }
}

// WithInactivityTimeout overrides the default 60-second inactivity timeout.
func WithInactivityTimeout(d time.Duration) SessionStoreOption {
	return func(s *SessionStore) { s.timeout = d }
}

// NewSessionStoreWithOptions creates a SessionStore with custom options.
func NewSessionStoreWithOptions(clk clock.Clock, opts ...SessionStoreOption) *SessionStore {
	s := NewSessionStore(clk)
	for _, o := range opts {
		o(s)
	}
	return s
}

// Lock acquires the session's field lock. See the [Session] type doc for the
// lock order relative to the store lock.
func (sess *Session) Lock() { sess.mu.Lock() }

// Unlock releases the session's field lock.
func (sess *Session) Unlock() { sess.mu.Unlock() }

// Allocate creates a new pending session and returns it.
// If capacity is reached, it evicts the oldest pending session (LRU per spec).
// Returns [ErrSessionFull] only when all slots are occupied by active sessions.
//
// maxPriv and channel are stored before the session is inserted into the map so
// the struct is fully initialized before it becomes reachable to other
// goroutines; callers must not write session fields after Allocate returns
// without holding the session lock.
func (s *SessionStore) Allocate(consoleID uint32, authAlg types.AuthAlg, integrityAlg types.IntegrityAlg, cryptAlg types.CryptAlg, maxPriv PrivilegeLevel, channel uint8) (*Session, error) {
	s.EvictExpired()

	s.mu.Lock()
	if len(s.sessions) >= s.max {
		// Evict oldest pending session if any exist. evictOldestPending takes
		// the store lock itself, so release it first to keep the lock order.
		s.mu.Unlock()
		if !s.evictOldestPending() {
			return nil, ErrSessionFull
		}
		s.mu.Lock()
		// Re-check capacity: evictOldestPending released the store lock, so a
		// concurrent Allocate could have refilled the slot it freed.
		if len(s.sessions) >= s.max {
			s.mu.Unlock()
			return nil, ErrSessionFull
		}
	}
	defer s.mu.Unlock()

	bmcID, err := randomUint32()
	if err != nil {
		return nil, fmt.Errorf("generate session ID: %w", err)
	}
	// Avoid collision with existing IDs.
	for s.sessions[bmcID] != nil || bmcID == 0 {
		bmcID, err = randomUint32()
		if err != nil {
			return nil, fmt.Errorf("generate session ID: %w", err)
		}
	}

	now := s.clock.Now()
	sess := &Session{
		BMCID:        bmcID,
		ConsoleID:    consoleID,
		State:        SessionStatePending,
		AuthAlg:      authAlg,
		IntegrityAlg: integrityAlg,
		CryptAlg:     cryptAlg,
		MaxPrivilege: maxPriv,
		Channel:      channel,
		CreatedAt:    now,
		LastActivity: now,
	}
	s.sessions[bmcID] = sess
	return sess, nil
}

// Get returns the session for bmcID, or [ErrNoSession]. It touches no session
// field; the caller updates [Session.LastActivity] under the session lock while
// dispatching the packet.
func (s *SessionStore) Get(bmcID uint32) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[bmcID]
	if !ok {
		return nil, fmt.Errorf("session 0x%08x: %w", bmcID, ErrNoSession)
	}
	return sess, nil
}

// Close marks a session as closed and removes it from the store.
func (s *SessionStore) Close(bmcID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[bmcID]; !ok {
		return fmt.Errorf("session 0x%08x: %w", bmcID, ErrNoSession)
	}
	delete(s.sessions, bmcID)
	return nil
}

// sessionEntry pairs a session with its current map key so eviction can delete
// by key without reading any session field under the store lock.
type sessionEntry struct {
	id   uint32
	sess *Session
}

// snapshot copies the current (id, session) pairs under the store lock. It
// reads no session fields, so it never needs a session lock.
func (s *SessionStore) snapshot() []sessionEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries := make([]sessionEntry, 0, len(s.sessions))
	for id, sess := range s.sessions {
		entries = append(entries, sessionEntry{id: id, sess: sess})
	}
	return entries
}

// deleteIfIdentity removes id under the store lock, but only if it still maps to
// the same session pointer. The caller holds the session's lock, so the
// eviction-condition re-check and this delete form one atomic step against
// concurrent refresh/activation of that session. Returns 1 if removed.
func (s *SessionStore) deleteIfIdentity(id uint32, sess *Session) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.sessions[id]; ok && cur == sess {
		delete(s.sessions, id)
		return 1
	}
	return 0
}

// EvictExpired removes all sessions that have been inactive beyond the timeout.
// Called periodically by the server and from [SessionStore.Allocate].
//
// It snapshots the session set under the store lock, then for each session tries
// its lock with TryLock. A session whose lock is busy (held by an in-flight
// packet, or by the current goroutine) is skipped this pass: eviction never
// blocks on a session lock, which is what would otherwise self-deadlock when a
// handler triggers eviction while holding its own session lock. With the lock
// held it re-checks expiry and deletes under the store lock (session-then-store
// order), so a session refreshed between snapshot and delete is never evicted.
func (s *SessionStore) EvictExpired() int {
	entries := s.snapshot()
	now := s.clock.Now()
	limit := s.timeout + DefaultInactivityTimeoutTolerance

	n := 0
	for _, e := range entries {
		if !e.sess.mu.TryLock() {
			continue
		}
		if now.Sub(e.sess.LastActivity) > limit {
			n += s.deleteIfIdentity(e.id, e.sess)
		}
		e.sess.mu.Unlock()
	}
	return n
}

// evictOldestPending removes the oldest pending session, returning true if one
// was removed. It follows the same snapshot / TryLock-per-session pattern as
// EvictExpired, then re-validates the chosen victim under its own lock before
// deleting it so an activation racing the scan is never clobbered.
func (s *SessionStore) evictOldestPending() bool {
	entries := s.snapshot()

	var oldest *sessionEntry
	var oldestCreated time.Time
	for i := range entries {
		e := entries[i]
		if !e.sess.mu.TryLock() {
			continue
		}
		pending := e.sess.State == SessionStatePending
		created := e.sess.CreatedAt
		e.sess.mu.Unlock()
		if pending && (oldest == nil || created.Before(oldestCreated)) {
			oldest = &entries[i]
			oldestCreated = created
		}
	}
	if oldest == nil {
		return false
	}
	if !oldest.sess.mu.TryLock() {
		return false
	}
	defer oldest.sess.mu.Unlock()
	if oldest.sess.State != SessionStatePending {
		return false
	}
	return s.deleteIfIdentity(oldest.id, oldest.sess) > 0
}

// Count returns the number of sessions currently in the store.
func (s *SessionStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

// Cap returns the maximum number of concurrent sessions the store can hold,
// i.e. the number of slots in the session table.
func (s *SessionStore) Cap() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max
}

// InboundSeqValid checks whether seq is within the acceptable sliding window
// defined by the IPMI spec (section 6.12.13):  +15 / -16 of the last accepted value.
// Session sequence numbers start at 1; 0 is reserved for pre-session packets.
func InboundSeqValid(last, seq uint32) bool {
	if seq == 0 {
		return false
	}
	diff := int64(seq) - int64(last)
	return diff >= -16 && diff <= 15
}

func randomUint32() (uint32, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}
