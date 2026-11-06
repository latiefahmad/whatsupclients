package wa

import (
	"bytes"
	"context"
	"strings"
	"testing"

	appstate "github.com/polymorfa/hypermeow/appstate"
	waBinary "github.com/polymorfa/hypermeow/binary"
	waServerSync "github.com/polymorfa/hypermeow/proto/waServerSync"
	waLog "github.com/polymorfa/hypermeow/util/log"
	"google.golang.org/protobuf/proto"
)

// appStateIQ builds an IQ response shaped like the server's answer to a
// snapshot request: <iq><sync><collection name="regular_high"/></sync></iq>.
func appStateIQ(name string) *waBinary.Node {
	collection := waBinary.Node{Tag: "collection", Attrs: waBinary.Attrs{"name": name}}
	return &waBinary.Node{
		Tag:   "iq",
		Attrs: waBinary.Attrs{"type": "result"},
		Content: []waBinary.Node{{
			Tag:     "sync",
			Content: []waBinary.Node{collection},
		}},
	}
}

// TestSnapshotDiagSilentWithoutDebug keeps the hook out of a normal run. It
// parses every app state IQ the server sends, which costs work hypermeow is
// about to do again.
func TestSnapshotDiagSilentWithoutDebug(t *testing.T) {
	log := &captureLogger{}
	b := &Backend{ctx: context.Background(), log: log}
	b.noteAppStateSnapshot(b.ctx, appStateIQ("regular_high"))
	for _, line := range log.lines {
		if strings.Contains(line, "app state diag") {
			t.Fatalf("logged without -debug: %q", line)
		}
	}
}

// TestSnapshotDiagLogsTheShape checks that an app state collection the
// server answers with is described in the log, snapshot or not: a collection
// this device holds no state for can only be seeded by a snapshot, so
// "no snapshot" is itself the answer to look for.
func TestSnapshotDiagLogsTheShape(t *testing.T) {
	log := &captureLogger{}
	b := &Backend{ctx: context.Background(), log: log, debugLog: true}
	b.noteAppStateSnapshot(b.ctx, appStateIQ("regular_high"))

	var line string
	for _, l := range log.lines {
		if strings.Contains(l, "app state diag: collection") {
			line = l
		}
	}
	if line == "" {
		t.Fatalf("no diagnostics line for an app state IQ: %q", log.lines)
	}
	if !strings.Contains(line, "regular_high") {
		t.Errorf("the line doesn't name the collection: %q", line)
	}
}

// appStateIQWithSnapshot builds the answer the server actually gives a
// snapshot request: the <snapshot> child holds an ExternalBlobReference, not
// the snapshot itself, and patches follow it.
func appStateIQWithSnapshot(t *testing.T, name string, fileSize uint64) *waBinary.Node {
	t.Helper()
	handle := "4:abc123"
	ref, err := proto.Marshal(&waServerSync.ExternalBlobReference{
		Handle:        &handle,
		FileSizeBytes: &fileSize,
		FileSHA256:    []byte("0123456789abcdef0123456789abcdef"),
	})
	if err != nil {
		t.Fatal(err)
	}
	version := uint64(305)
	patch, err := proto.Marshal(&waServerSync.SyncdPatch{Version: &waServerSync.SyncdVersion{Version: &version}})
	if err != nil {
		t.Fatal(err)
	}
	return &waBinary.Node{
		Tag: "iq",
		Content: []waBinary.Node{{
			Tag: "sync",
			Content: []waBinary.Node{{
				Tag:   "collection",
				Attrs: waBinary.Attrs{"name": name},
				Content: []waBinary.Node{
					{Tag: "snapshot", Content: ref},
					{Tag: "patches", Content: []waBinary.Node{
						{Tag: "patch", Content: patch},
					}},
				},
			}},
		}},
	}
}

// TestSnapshotDiagReadsTheReference checks the line a refused snapshot is
// explained by: how big the blob it points at is, and which patches came with
// it. The 308 bytes the reference itself takes say nothing about the state.
func TestSnapshotDiagReadsTheReference(t *testing.T) {
	log := &captureLogger{}
	b := &Backend{ctx: context.Background(), log: log, debugLog: true}
	b.noteAppStateSnapshot(b.ctx, appStateIQWithSnapshot(t, "regular_high", 91234))

	var sawRef, sawPatch bool
	for _, l := range log.lines {
		if strings.Contains(l, "filesize=91234") && strings.Contains(l, "handle=4:abc123") {
			sawRef = true
		}
		if strings.Contains(l, "patch v305") {
			sawPatch = true
		}
	}
	if !sawRef {
		t.Errorf("the snapshot reference wasn't logged: %q", log.lines)
	}
	if !sawPatch {
		t.Errorf("the patches after the snapshot weren't logged: %q", log.lines)
	}
}

// TestSnapshotDiagSaysWhenThereIsNone covers the other answer that would
// explain a collection never seeding: the server sends patches only.
func TestSnapshotDiagSaysWhenThereIsNone(t *testing.T) {
	log := &captureLogger{}
	b := &Backend{ctx: context.Background(), log: log, debugLog: true}
	b.noteAppStateSnapshot(b.ctx, appStateIQ("regular_high"))

	var sawMissing bool
	for _, l := range log.lines {
		if strings.Contains(l, "no snapshot at all") {
			sawMissing = true
		}
	}
	if !sawMissing {
		t.Errorf("a collection with no snapshot wasn't reported: %q", log.lines)
	}
}

// TestSnapshotMACFollowsTheVersionAndName checks the MAC is bound to both,
// which is what makes a mismatch meaningful: a snapshot sealed for another
// version or collection can't verify here even with the right key.
func TestSnapshotMACFollowsTheVersionAndName(t *testing.T) {
	expanded := expandKey(bytes.Repeat([]byte{7}, 32))
	var hash [128]byte
	base := snapshotMAC(hash, 305, appstate.WAPatchRegularHigh, expanded)

	if other := snapshotMAC(hash, 306, appstate.WAPatchRegularHigh, expanded); bytes.Equal(base, other) {
		t.Error("the same hash signed at two versions gave one MAC")
	}
	if other := snapshotMAC(hash, 305, appstate.WAPatchRegularLow, expanded); bytes.Equal(base, other) {
		t.Error("the same hash signed for two collections gave one MAC")
	}
	if len(base) != 32 {
		t.Errorf("a snapshot MAC is 32 bytes, got %d", len(base))
	}
}

// TestExpandKeyLayout pins the subkey split the MAC is taken from. The
// snapshot key is the fourth of five 32-byte subkeys; taking another one
// silently produced a MAC that never matches.
func TestExpandKeyLayout(t *testing.T) {
	expanded := expandKey(bytes.Repeat([]byte{7}, 32))
	if len(expanded) != 160 {
		t.Fatalf("five 32-byte subkeys, got %d bytes", len(expanded))
	}
	other := expandKey(bytes.Repeat([]byte{8}, 32))
	if bytes.Equal(expanded[96:128], other[96:128]) {
		t.Error("the snapshot key doesn't depend on the key data")
	}
}

// TestSnapshotDiagLeavesTheNodeAlone is the point of a handler running on
// hypermeow's dispatch goroutine: the node it sees must reach hypermeow
// unchanged, or verification and decoding see something else.
func TestSnapshotDiagLeavesTheNodeAlone(t *testing.T) {
	b := &Backend{ctx: context.Background(), log: waLog.Noop, debugLog: true}
	node := appStateIQ("regular_high")
	before := node.String()
	b.noteAppStateSnapshot(b.ctx, node)
	if after := node.String(); after != before {
		t.Fatalf("diagnostics changed the node:\nbefore %s\nafter  %s", before, after)
	}
}
