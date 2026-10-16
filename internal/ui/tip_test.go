package ui

import (
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/latiefahmad/whatsupclients/internal/mock"
)

// TestTipDwell checks the tooltip state machine: nothing without hover,
// registration on hover, a fresh timer per button, and clearing.
func TestTipDwell(t *testing.T) {
	u := New(mock.New())
	var ops op.Ops
	frame := func(now time.Time) layout.Context {
		ops.Reset()
		return layout.Context{Ops: &ops, Now: now,
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1100, 700))}
	}
	now := testNow()
	u.tip(frame(now), "a", "Archive", false)
	if u.tipKey != "" {
		t.Fatal("a tooltip without hover")
	}
	u.tip(frame(now), "a", "Archive", true)
	if u.tipKey != "a" || u.tipText != "Archive" {
		t.Fatalf("tooltip %q %q", u.tipKey, u.tipText)
	}
	since := u.tipSince
	u.tip(frame(now.Add(100*time.Millisecond)), "a", "Archive", true)
	if !u.tipSince.Equal(since) {
		t.Fatal("same-button hover restarted the dwell")
	}
	u.tip(frame(now), "b", "Mute", true)
	if u.tipKey != "b" || !u.tipSince.Equal(now) {
		t.Fatal("another button didn't restart the dwell")
	}
	u.tip(frame(now), "b", "Mute", false)
	if u.tipKey != "" {
		t.Fatal("leaving didn't clear the tooltip")
	}
	// Drawing below a button before the dwell draws nothing.
	u.tip(frame(now), "a", "Archive", true)
	u.tipBelow(frame(now.Add(100*time.Millisecond)), "a", "Archive", 38, 44, true)
	u.tipBelow(frame(now.Add(time.Second)), "a", "Archive", 38, 44, true)
	u.tipBelow(frame(now), "a", "Archive", 38, 44, false)
	if u.tipKey != "" {
		t.Fatal("leaving didn't clear the tooltip")
	}
}
