package wa

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/polymorfa/hypermeow/appstate"
	"github.com/polymorfa/hypermeow/proto/waSyncAction"
	"google.golang.org/protobuf/proto"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

func deletePatch(chat, id string) appstate.PatchInfo {
	return appstate.PatchInfo{
		Type: appstate.WAPatchRegularHigh,
		Mutations: []appstate.MutationInfo{{
			Index:   []string{appstate.IndexDeleteMessageForMe, chat, id, "0", "0"},
			Version: 3,
			Value: &waSyncAction.SyncActionValue{DeleteMessageForMeAction: &waSyncAction.DeleteMessageForMeAction{
				DeleteMedia:      proto.Bool(true),
				MessageTimestamp: proto.Int64(1700000000),
			}},
		}},
	}
}

// TestPatchKeptWhenSendFails checks that a change the phone refused is kept,
// so it can be sent again: before this it was dropped, and the message stayed
// deleted here while the phone kept it.
func TestPatchKeptWhenSendFails(t *testing.T) {
	b := newQueueTestBackend(t)
	b.enqueuePatch(deletePatch("c@s.whatsapp.net", "m1"))

	if n := b.pendingLen(); n != 1 {
		t.Fatalf("pending: %d, want 1", n)
	}
}

// TestQueueStopsGrowingAtTheLimit checks the bound that keeps a collection
// which can never accept a patch from growing for the whole session. The
// newest changes are the ones kept: they match what the chat shows now.
func TestQueueStopsGrowingAtTheLimit(t *testing.T) {
	b := newQueueTestBackend(t)
	for i := 0; i < maxPendingPerCollection+25; i++ {
		b.enqueuePatch(deletePatch("c@s.whatsapp.net", fmt.Sprintf("m%d", i)))
	}
	if n := b.pendingLen(); n != maxPendingPerCollection {
		t.Fatalf("pending: %d, want the limit %d", n, maxPendingPerCollection)
	}
	// The oldest went, the newest stayed.
	b.patches.mu.Lock()
	first := b.patches.pending[0].Index[2]
	last := b.patches.pending[len(b.patches.pending)-1].Index[2]
	b.patches.mu.Unlock()
	if first == "m0" {
		t.Error("the oldest change was kept instead of dropped")
	}
	if want := fmt.Sprintf("m%d", maxPendingPerCollection+24); last != want {
		t.Errorf("newest kept is %s, want %s", last, want)
	}
}

// TestQueueFullStopsTheRetry checks the signal the send path uses to stop
// asking: once a collection is full, nothing more can be done for it, so a
// resend would only log the same warning again.
func TestQueueFullStopsTheRetry(t *testing.T) {
	b := newQueueTestBackend(t)
	if b.queueFull(string(appstate.WAPatchRegularHigh)) {
		t.Fatal("an empty queue reported full")
	}
	for i := 0; i < maxPendingPerCollection; i++ {
		b.enqueuePatch(deletePatch("c@s.whatsapp.net", fmt.Sprintf("m%d", i)))
	}
	if !b.queueFull(string(appstate.WAPatchRegularHigh)) {
		t.Fatal("a full queue didn't report full")
	}
	// Another collection is unaffected.
	if b.queueFull(string(appstate.WAPatchRegularLow)) {
		t.Fatal("a different collection reported full")
	}
}

// TestTrimLeavesOtherCollectionsAlone checks the bound is per collection, so a
// stuck one can't push out the changes of a healthy one.
func TestTrimLeavesOtherCollectionsAlone(t *testing.T) {
	b := newQueueTestBackend(t)
	low := deletePatch("c@s.whatsapp.net", "low1")
	low.Type = appstate.WAPatchRegularLow
	b.enqueuePatch(low)
	for i := 0; i < maxPendingPerCollection+5; i++ {
		b.enqueuePatch(deletePatch("c@s.whatsapp.net", fmt.Sprintf("m%d", i)))
	}
	b.patches.mu.Lock()
	defer b.patches.mu.Unlock()
	for _, q := range b.patches.pending {
		if q.Name == string(appstate.WAPatchRegularLow) && q.Index[2] == "low1" {
			return
		}
	}
	t.Fatal("a healthy collection's change was dropped")
}

// TestPatchDroppedWhenSent checks that a change that reached the phone isn't
// kept, or it would be sent twice.
func TestPatchDroppedWhenSent(t *testing.T) {
	b := newQueueTestBackend(t)
	p := deletePatch("c@s.whatsapp.net", "m1")
	b.enqueuePatch(p)
	b.forgetPatch(p)

	if n := b.pendingLen(); n != 0 {
		t.Errorf("pending: %d, want 0", n)
	}
}

// TestPatchQueueSurvivesRestart checks that a change kept before the app
// closed is still there when it opens again.
func TestPatchQueueSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := msgStore{db: db}
	if err := s.init(ctx); err != nil {
		t.Fatal(err)
	}
	first := &Backend{store: s, ctx: ctx, log: testLogger{}}
	first.enqueuePatch(deletePatch("c@s.whatsapp.net", "m1"))

	second := &Backend{store: s, ctx: ctx, log: testLogger{}}
	second.loadPatches()
	if n := second.pendingLen(); n != 1 {
		t.Fatalf("after restart pending: %d, want 1", n)
	}
	if got := second.patches.pending[0].Index[2]; got != "m1" {
		t.Errorf("message id: %q, want m1", got)
	}
}

// TestPatchNotQueuedTwice checks that the same change failing twice is kept
// once, not twice.
func TestPatchNotQueuedTwice(t *testing.T) {
	b := newQueueTestBackend(t)
	p := deletePatch("c@s.whatsapp.net", "m1")
	b.enqueuePatch(p)
	b.enqueuePatch(p)

	if n := b.pendingLen(); n != 1 {
		t.Errorf("pending: %d, want 1", n)
	}
}

// TestPatchValueRoundTrip checks that a kept change keeps what it does, not
// just its key: a delete must still be a delete when it goes out again.
func TestPatchValueRoundTrip(t *testing.T) {
	b := newQueueTestBackend(t)
	b.enqueuePatch(deletePatch("c@s.whatsapp.net", "m1"))
	p := b.patches.pending[0]
	if len(p.Value) == 0 {
		t.Fatal("value not kept")
	}
	v := &waSyncAction.SyncActionValue{}
	if err := proto.Unmarshal(p.Value, v); err != nil {
		t.Fatalf("value doesn't decode: %v", err)
	}
	if v.GetDeleteMessageForMeAction() == nil {
		t.Error("kept value is no longer a delete")
	}
	if v.GetDeleteMessageForMeAction().GetMessageTimestamp() != 1700000000 {
		t.Error("kept value lost its timestamp")
	}
}

func newQueueTestBackend(t *testing.T) *Backend {
	t.Helper()
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	s := msgStore{db: db}
	if err := s.init(ctx); err != nil {
		t.Fatal(err)
	}
	return &Backend{store: s, ctx: ctx, log: testLogger{}}
}

func (b *Backend) pendingLen() int {
	b.patches.mu.Lock()
	defer b.patches.mu.Unlock()
	return len(b.patches.pending)
}

var _ = model.KindText
