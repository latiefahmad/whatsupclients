package ui

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/latiefahmad/whatsupclients/internal/mock"
)

// TestStatusViews checks the "Viewed by" sheet over your own status: the
// eye opens it, its rows open the person's chat, and Esc closes the
// sheet before the viewer.
func TestStatusViews(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.applyEvents()
	u.setPage(pageStatus)
	now := testNow()
	var ops op.Ops
	var r input.Router
	frame := func() {
		for range 3 {
			ops.Reset()
			u.Layout(layout.Context{Ops: &ops, Now: now, Source: r.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
				Constraints: layout.Exact(image.Pt(1100, 700))})
			r.Frame(&ops)
			now = now.Add(50 * time.Millisecond)
		}
		u.applyEvents()
	}
	frame()

	mine := u.myStatus()
	if mine == nil || len(mine.Updates) == 0 || mine.Updates[0].Views != 4 {
		t.Fatalf("my status = %+v, want 4 views", mine)
	}
	u.status.viewer.show(mine)
	u.status.viewer.anim.snap(true)
	frame()
	if u.status.viewer.views.open {
		t.Error("the sheet is open before its eye was clicked")
	}
	// Clicking the eye opens the sheet with the demo viewers.
	u.status.viewer.views.btn.Click()
	frame()
	if !u.status.viewer.views.open {
		t.Fatal("the eye didn't open the sheet")
	}
	v := u.backend.StatusViews(mine.Updates[0].ID)
	if len(v.Viewers) != 4 || len(v.Replies) != 1 || v.Replies[0].Msg.ChatID != "citra" {
		t.Fatalf("views = %+v", v)
	}
	// Esc closes the sheet first, the viewer only then.
	u.Escape()
	frame()
	if u.status.viewer.views.open {
		t.Fatal("Esc didn't close the sheet")
	}
	if !u.status.viewer.isOpen() {
		t.Fatal("Esc closed the viewer with the sheet")
	}
	// The reply's row opens its chat, and the viewer closes with it.
	u.status.viewer.views.btn.Click()
	frame()
	if !u.status.viewer.views.open {
		t.Fatal("the sheet didn't reopen")
	}
	u.btn("st:reply:0").Click()
	frame()
	if u.selected == nil || u.selected.ID != "citra" {
		t.Fatalf("chat after clicking the reply = %v", u.selected)
	}
	if u.status.viewer.views.open || u.status.viewer.isOpen() {
		t.Error("the viewer stayed open over the chat")
	}
}
