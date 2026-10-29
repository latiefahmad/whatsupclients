package wa

import (
	"context"
	"testing"
	"time"
)

// TestResyncAppStateIsRateLimited checks that a broken collection isn't
// refetched on every sync error. Each failure dispatches another
// AppStateSyncError, so an unthrottled resync loops twice a second until
// WhatsApp answers "429 rate-overlimit", and nothing the collection carries
// (a message deleted for me, a cleared chat) is applied meanwhile.
func TestResyncAppStateIsRateLimited(t *testing.T) {
	ctx := context.Background()
	s, _ := testStore(t)
	b := &Backend{ctx: ctx, store: *s}
	const key = "appstate_resync:regular_high"

	if !b.takeResyncSlot(key) {
		t.Fatal("the first resync was refused")
	}
	for i := 0; i < 5; i++ {
		if b.takeResyncSlot(key) {
			t.Fatalf("resync %d ran inside the gap: five errors in a row would mean five requests", i+2)
		}
	}

	// Outside the gap it runs again, so a collection that stays broken is
	// still retried.
	_ = s.setMetaValue(ctx, key, time.Now().Add(-2*resyncGap).Format(time.RFC3339))
	if !b.takeResyncSlot(key) {
		t.Error("no resync ran after the gap; a broken collection would never retry")
	}

	// A different collection has a slot of its own.
	if !b.takeResyncSlot("appstate_resync:critical_block") {
		t.Error("another collection was throttled by this one's marker")
	}
}

// TestResyncGapIsWellUnderRecoveryGap keeps the two throttles apart: the
// collection is refetched a few times a minute at worst, while the phone is
// asked to repair it far less often.
func TestResyncGapIsWellUnderRecoveryGap(t *testing.T) {
	if resyncGap <= 0 {
		t.Fatalf("resyncGap %v would allow a tight loop", resyncGap)
	}
	if resyncGap >= recoveryGap {
		t.Errorf("resyncGap %v should stay well under recoveryGap %v", resyncGap, recoveryGap)
	}
}
