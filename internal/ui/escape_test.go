package ui

import "testing"

// TestEscapeClosesChat checks that Esc, once nothing above it is open,
// closes the info panel first and then the chat, and that the chat's
// typed draft waits for its next open.
func TestEscapeClosesChat(t *testing.T) {
	u, _, frame, _ := chatListUI(t)
	u.SelectID("citra")
	for range 2 {
		frame()
	}
	if u.selected == nil {
		t.Fatal("citra isn't open")
	}

	// The info panel closes first, the chat stays.
	u.openInfo("citra")
	frame()
	if !u.info.open {
		t.Fatal("the info panel didn't open")
	}
	u.Escape()
	frame()
	if u.info.open {
		t.Error("Esc didn't close the info panel")
	}
	if u.selected == nil {
		t.Fatal("Esc closed the chat from the info panel")
	}

	// Then the chat closes, its draft stashed along.
	u.conv.composer.SetText("half a thought")
	u.Escape()
	frame()
	if u.selected != nil {
		t.Fatal("Esc didn't close the chat")
	}
	if d := u.drafts["citra"]; d == nil || d.text != "half a thought" {
		t.Fatalf("the draft after Esc is %+v", d)
	}

	// Reopening gives it back.
	u.SelectID("citra")
	frame()
	if got := u.conv.composer.Text(); got != "half a thought" {
		t.Fatalf("reopening shows %q", got)
	}
	if u.drafts["citra"] != nil {
		t.Fatal("the reopened chat still has a draft")
	}
}
