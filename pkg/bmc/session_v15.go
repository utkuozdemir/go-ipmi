package bmc

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
)

// V15AuthType mirrors IPMI v1.5 authentication type codes.
type V15AuthType uint8

const (
	V15AuthTypeNone     V15AuthType = 0x00
	V15AuthTypeMD2      V15AuthType = 0x01
	V15AuthTypeMD5      V15AuthType = 0x02
	V15AuthTypePassword V15AuthType = 0x04
	V15AuthTypeOEM      V15AuthType = 0x05
)

// v15InboundWindow is the inbound sequence sliding window (spec v1.5§6.11.11 Option 1 / v2.0§6.12.11 Option 1).
const v15InboundWindow = 8

// DefaultV15AuthTypes is the default set of v1.5 auth types the reference BMC
// advertises and accepts.
var DefaultV15AuthTypes = []V15AuthType{V15AuthTypeMD5}

// V15AuthTypeToCapsBit maps an auth type to the corresponding bit in Get
// Channel Authentication Capabilities response byte 3 (bits [5:0]).
func V15AuthTypeToCapsBit(t V15AuthType) uint8 {
	switch t {
	case V15AuthTypeNone:
		return 1 << 0
	case V15AuthTypeMD2:
		return 1 << 1
	case V15AuthTypeMD5:
		return 1 << 2
	case V15AuthTypePassword:
		return 1 << 4
	case V15AuthTypeOEM:
		return 1 << 5
	default:
		return 0
	}
}

// V15SessionState tracks IPMI v1.5 session negotiation progress.
type V15SessionState uint8

const (
	V15SessionStatePending V15SessionState = iota
	V15SessionStateActive
	V15SessionStateClosed
)

// V15Session holds IPMI v1.5 session state.
//
// Concurrency: the session mutex guards the per-packet fields written during
// dispatch: the inbound/outbound sequence counters, InboundRcvd, and writes to
// LastActivity. State, User, MaxPrivilege and PrivilegeLevel are written only by
// [V15SessionStore.Activate], which runs under the store lock while its caller
// holds the session lock, so those fields are covered by both locks; the Count*
// helpers read them under the store lock, and eviction reads State, CreatedAt and
// LastActivity under the session lock (via TryLock). TempSessionID, SessionID,
// AuthType, Challenge, Channel and CreatedAt are set before the session is
// published. Lock order is always session-then-store.
type V15Session struct {
	// mu serializes access to all fields below.
	mu sync.Mutex

	TempSessionID uint32
	SessionID     uint32
	State         V15SessionState

	AuthType  V15AuthType
	Challenge [16]byte

	InboundSeq  uint32
	InboundRcvd uint8 // bitmap: bit i => (InboundSeq - i) received
	OutboundSeq uint32

	User           *User
	PrivilegeLevel PrivilegeLevel
	MaxPrivilege   PrivilegeLevel

	Channel uint8

	CreatedAt    time.Time
	LastActivity time.Time
}

// V15SessionStore is a thread-safe registry of IPMI v1.5 sessions.
type V15SessionStore struct {
	mu       sync.Mutex
	sessions map[uint32]*V15Session
	max      int
	timeout  time.Duration
	clock    clock.Clock
}

// NewV15SessionStore creates a V15SessionStore with the default limits.
func NewV15SessionStore(clk clock.Clock) *V15SessionStore {
	return &V15SessionStore{
		sessions: make(map[uint32]*V15Session, MaxSessions),
		max:      MaxSessions,
		timeout:  DefaultInactivityTimeout,
		clock:    clk,
	}
}

// Lock acquires the session's field lock. See the [Session] type doc for the
// lock order relative to the store lock.
func (sess *V15Session) Lock() { sess.mu.Lock() }

// Unlock releases the session's field lock.
func (sess *V15Session) Unlock() { sess.mu.Unlock() }

// CreatePending allocates a pending v1.5 session after Get Session Challenge.
// The session is fully initialized before it is inserted into the map.
func (s *V15SessionStore) CreatePending(authType V15AuthType, user *User, challenge [16]byte, channel uint8) (*V15Session, error) {
	s.EvictExpired()

	s.mu.Lock()
	if len(s.sessions) >= s.max {
		s.mu.Unlock()
		if !s.evictOldestPending() {
			return nil, ErrSessionFull
		}
		s.mu.Lock()
		// Re-check capacity: evictOldestPending released the store lock, so a
		// concurrent CreatePending could have refilled the slot it freed.
		if len(s.sessions) >= s.max {
			s.mu.Unlock()
			return nil, ErrSessionFull
		}
	}
	defer s.mu.Unlock()

	tempID, err := randomUint32()
	if err != nil {
		return nil, fmt.Errorf("generate temp session ID: %w", err)
	}
	for s.sessions[tempID] != nil || tempID == 0 {
		tempID, err = randomUint32()
		if err != nil {
			return nil, fmt.Errorf("generate temp session ID: %w", err)
		}
	}

	now := s.clock.Now()
	sess := &V15Session{
		TempSessionID: tempID,
		State:         V15SessionStatePending,
		AuthType:      authType,
		Challenge:     challenge,
		User:          user,
		Channel:       channel,
		CreatedAt:     now,
		LastActivity:  now,
	}
	s.sessions[tempID] = sess
	return sess, nil
}

// Get returns a session by its current lookup ID without updating activity.
func (s *V15SessionStore) Get(id uint32) (*V15Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("v1.5 session 0x%08x: %w", id, ErrNoSession)
	}
	return sess, nil
}

// Cap returns the maximum number of concurrent v1.5 sessions the store can
// hold, i.e. the number of slots in the session table.
func (s *V15SessionStore) Cap() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max
}

// CountActiveSessions returns the number of active v1.5 sessions.
func (s *V15SessionStore) CountActiveSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, sess := range s.sessions {
		if sess.State == V15SessionStateActive {
			n++
		}
	}
	return n
}

// CountActiveSessionsForUser returns active sessions owned by userID.
func (s *V15SessionStore) CountActiveSessionsForUser(userID uint8) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, sess := range s.sessions {
		if sess.State == V15SessionStateActive && sess.User != nil && sess.User.ID == userID {
			n++
		}
	}
	return n
}

// CountActiveSessionsWithMaxPrivilegeAtLeast counts active sessions whose
// negotiated maximum privilege is >= min (for Table 18-17 completion 0x83).
func (s *V15SessionStore) CountActiveSessionsWithMaxPrivilegeAtLeast(min PrivilegeLevel) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, sess := range s.sessions {
		if sess.State == V15SessionStateActive && sess.MaxPrivilege >= min {
			n++
		}
	}
	return n
}

// Activate transitions a pending session to active with a new permanent ID.
// maxPrivilege is the requested ceiling; initial privilege is USER per v1.5§6.8 / v2.0§6.8
// (Callback when max is Callback).
//
// inboundSeq is the Activate Session response "Session inbound sequence number"
// (spec v1.5§18.15 / v2.0§6.12.9): the starting sequence the remote console must use on
// its first authenticated packet. InboundSeq on the session tracks the highest
// sequence already accepted, so it is seeded to inboundSeq-1 (wrapping) with an
// empty receive bitmap — otherwise the first packet (seq == inboundSeq) is
// rejected as a duplicate and clients such as ipmitool stall for a full LAN
// timeout before retrying with inboundSeq+1.
//
// Precondition: the caller must hold pending's session lock. Activate mutates
// fields of an already-published session (State, User's derived privilege, seq
// counters, LastActivity) while holding only the store lock; it is safe only
// because the sole caller holds the session lock during dispatch, so the lock
// order stays session-then-store and no reader observes a half-updated session.
func (s *V15SessionStore) Activate(pending *V15Session, permanentID, inboundSeq, outboundSeq uint32, maxPrivilege PrivilegeLevel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pending.State != V15SessionStatePending {
		return errors.New("session is not pending")
	}
	delete(s.sessions, pending.TempSessionID)

	for s.sessions[permanentID] != nil || permanentID == 0 {
		var err error
		permanentID, err = randomUint32()
		if err != nil {
			return fmt.Errorf("generate permanent session ID: %w", err)
		}
	}

	initialPriv := PrivilegeLevelUser
	if maxPrivilege == PrivilegeLevelCallback {
		initialPriv = PrivilegeLevelCallback
	}

	pending.SessionID = permanentID
	pending.State = V15SessionStateActive
	// Seq 0 is reserved for pre-session; never advertise/seed it as a start.
	if inboundSeq == 0 {
		inboundSeq = 1
	}
	pending.InboundSeq = inboundSeq - 1
	pending.InboundRcvd = 0
	pending.OutboundSeq = outboundSeq
	pending.PrivilegeLevel = initialPriv
	pending.MaxPrivilege = maxPrivilege
	pending.LastActivity = s.clock.Now()

	s.sessions[permanentID] = pending
	return nil
}

// Close removes a session by permanent or temp ID.
func (s *V15SessionStore) Close(id uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return fmt.Errorf("v1.5 session 0x%08x: %w", id, ErrNoSession)
	}
	delete(s.sessions, id)
	if sess.State == V15SessionStateActive && sess.TempSessionID != id && sess.TempSessionID != 0 {
		delete(s.sessions, sess.TempSessionID)
	}
	return nil
}

// v15SessionEntry pairs a session with its current map key so eviction can
// delete by key without reading any session field under the store lock.
type v15SessionEntry struct {
	id   uint32
	sess *V15Session
}

func (s *V15SessionStore) snapshot() []v15SessionEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries := make([]v15SessionEntry, 0, len(s.sessions))
	for id, sess := range s.sessions {
		entries = append(entries, v15SessionEntry{id: id, sess: sess})
	}
	return entries
}

// deleteIfIdentity removes id under the store lock, but only if it still maps to
// the same session pointer. See [SessionStore.deleteIfIdentity] for why the
// caller holds the session lock across the re-check and this delete.
func (s *V15SessionStore) deleteIfIdentity(id uint32, sess *V15Session) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.sessions[id]; ok && cur == sess {
		delete(s.sessions, id)
		return 1
	}
	return 0
}

// EvictExpired removes inactive v1.5 sessions past the timeout. It tries each
// session's lock with TryLock and skips any busy one, so it never blocks on a
// session lock (which would self-deadlock when Get Session Challenge triggers
// eviction while holding its own session lock). See [SessionStore.EvictExpired]
// for the snapshot / TryLock / delete shape.
func (s *V15SessionStore) EvictExpired() int {
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

func (s *V15SessionStore) evictOldestPending() bool {
	entries := s.snapshot()

	var oldest *v15SessionEntry
	var oldestCreated time.Time
	for i := range entries {
		e := entries[i]
		if !e.sess.mu.TryLock() {
			continue
		}
		pending := e.sess.State == V15SessionStatePending
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
	if oldest.sess.State != V15SessionStatePending {
		return false
	}
	return s.deleteIfIdentity(oldest.id, oldest.sess) > 0
}

// v15SeqDiff returns seq-high as a signed delta with uint32 wrap-around.
func v15SeqDiff(high, seq uint32) int64 {
	diff := int64(seq) - int64(high)
	if diff > 1<<31 {
		diff -= 1 << 32
	} else if diff < -(1 << 31) {
		diff += 1 << 32
	}
	return diff
}

// TryAcceptInboundSeq implements spec v1.5§6.11.11 Option 1 / v2.0§6.12.11 Option 1 (+/-8 window, no dupes).
func (sess *V15Session) TryAcceptInboundSeq(seq uint32) bool {
	if seq == 0 {
		return false
	}
	high := sess.InboundSeq
	diff := v15SeqDiff(high, seq)

	if diff == 0 {
		return false
	}
	if diff > v15InboundWindow || diff < -v15InboundWindow {
		return false
	}

	if diff > 0 {
		shift := uint(diff)
		if shift > v15InboundWindow {
			return false
		}
		sess.InboundRcvd <<= shift
		sess.InboundRcvd |= 1
		sess.InboundSeq = seq
		return true
	}

	behind := uint(-diff)
	bit := uint8(1) << (behind - 1)
	if sess.InboundRcvd&bit != 0 {
		return false
	}
	sess.InboundRcvd |= bit
	return true
}

// V15InboundSeqValid reports whether seq is acceptable under Option 1 without
// mutating session state (for tests).
func V15InboundSeqValid(sess *V15Session, seq uint32) bool {
	if sess == nil || seq == 0 {
		return false
	}
	high := sess.InboundSeq
	diff := v15SeqDiff(high, seq)
	if diff == 0 {
		return false
	}
	if diff > v15InboundWindow || diff < -v15InboundWindow {
		return false
	}
	if diff < 0 {
		bit := uint8(1) << (uint(-diff) - 1)
		return sess.InboundRcvd&bit == 0
	}
	return true
}

// NextOutboundSeq returns the sequence number for the current outbound message
// and advances the counter for the next one. Sequence 0 is reserved for
// pre-session packets and is skipped on wrap (v1.5§6.11.9 / v2.0§6.12.9).
func (sess *V15Session) NextOutboundSeq() uint32 {
	if sess.OutboundSeq == 0 {
		sess.OutboundSeq = 1
	}
	seq := sess.OutboundSeq
	sess.OutboundSeq++
	// fix overflow
	if sess.OutboundSeq == 0 {
		sess.OutboundSeq = 1
	}
	return seq
}

// GenerateChallenge fills dst with random bytes for Get Session Challenge.
func GenerateChallenge(dst *[16]byte) error {
	_, err := rand.Read(dst[:])
	return err
}

// GenerateInboundSeq returns a non-zero initial inbound sequence number.
func GenerateInboundSeq() (uint32, error) {
	seq, err := randomUint32()
	if err != nil {
		return 0, err
	}
	if seq == 0 {
		seq = 1
	}
	return seq, nil
}

// PackSessionIDLE is a helper for auth code input construction.
func PackSessionIDLE(id uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, id)
	return b
}
