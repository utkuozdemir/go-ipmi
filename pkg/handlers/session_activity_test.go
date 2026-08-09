package handlers

import (
	"context"
	"encoding/binary"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// advanceableClock is a Clock whose current time can be moved forward by tests.
type advanceableClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *advanceableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *advanceableClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func (c *advanceableClock) NewTimer(d time.Duration) clock.Timer   { return clock.Real.NewTimer(d) }
func (c *advanceableClock) NewTicker(d time.Duration) clock.Ticker { return clock.Real.NewTicker(d) }

func newAdminBMC(clk clock.Clock) (*bmc.BMC, error) {
	info := bmc.DeviceInfo{
		DeviceID:                1,
		DeviceRevision:          1,
		FirmwareMajor:           2,
		IPMIVersion:             0x20,
		ManufacturerID:          0x000157,
		ProductID:               0x0001,
		AdditionalDeviceSupport: 0x39,
	}
	b := bmc.New(info, [16]byte{}, mock.New(), bmc.WithClock(clk))
	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		return nil, err
	}
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[lanChannelNumber] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}
	return b, nil
}

// TestHandleRAKP1RefreshesLastActivity asserts RAKP Message 1 bumps the
// session's LastActivity, so a session handshaking near the inactivity limit is
// not reaped despite receiving real traffic. RAKP runs outside the server's
// per-packet dispatch, which is the only other place LastActivity is refreshed.
func TestHandleRAKP1RefreshesLastActivity(t *testing.T) {
	clk := &advanceableClock{now: time.Now()}
	b, err := newAdminBMC(clk)
	if err != nil {
		t.Fatal(err)
	}

	sess, err := b.Sessions.Allocate(0x01020304, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128, bmc.PrivilegeLevelAdministrator, lanChannelNumber)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	created := sess.LastActivity

	clk.advance(30 * time.Second)

	resp, err := HandleRAKP1(context.Background(), b, rakp1Payload(sess.BMCID, bmc.PrivilegeLevelAdministrator, "ADMIN"))
	if err != nil {
		t.Fatalf("HandleRAKP1: %v", err)
	}
	if len(resp) < 2 || resp[1] != 0x00 {
		t.Fatalf("want successful RAKP2 response, got %x", resp)
	}
	if !sess.LastActivity.After(created) {
		t.Fatalf("LastActivity not refreshed: created=%v after=%v", created, sess.LastActivity)
	}
	if !sess.LastActivity.Equal(clk.Now()) {
		t.Fatalf("LastActivity=%v, want current clock %v", sess.LastActivity, clk.Now())
	}
}

// TestHandleRAKP3RefreshesLastActivity asserts RAKP Message 3 bumps LastActivity
// even on the failure path (the refresh happens under the session lock before any
// early return), matching the RAKP1 behavior above.
func TestHandleRAKP3RefreshesLastActivity(t *testing.T) {
	clk := &advanceableClock{now: time.Now()}
	b, err := newAdminBMC(clk)
	if err != nil {
		t.Fatal(err)
	}

	sess, err := b.Sessions.Allocate(0x01020304, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128, bmc.PrivilegeLevelAdministrator, lanChannelNumber)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	created := sess.LastActivity

	clk.advance(30 * time.Second)

	// Minimal RAKP3 payload: tag, statusCode=0, bmc session ID. The HMAC check
	// will fail, but LastActivity is refreshed before that.
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[4:8], sess.BMCID)
	if _, err := HandleRAKP3(context.Background(), b, payload); err != nil {
		t.Fatalf("HandleRAKP3: %v", err)
	}
	if !sess.LastActivity.After(created) {
		t.Fatalf("LastActivity not refreshed: created=%v after=%v", created, sess.LastActivity)
	}
}
