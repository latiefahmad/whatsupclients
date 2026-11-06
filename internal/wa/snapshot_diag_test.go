package wa

import (
	"context"
	"strings"
	"testing"

	waBinary "github.com/polymorfa/hypermeow/binary"
	waLog "github.com/polymorfa/hypermeow/util/log"
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
