package ui

import (
	"image"
	"slices"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// chatListUI is a demo window with frame and click helpers.
func chatListUI(t *testing.T) (*UI, *mock.Backend, func(), func(string)) {
	t.Helper()
	b := mock.New()
	u := New(b)
	u.Start(func() {})
	now := testNow()
	var ops op.Ops
	var r input.Router
	frame := func() {
		ops.Reset()
		u.Layout(layout.Context{Ops: &ops, Now: now, Source: r.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1600, 900))})
		r.Frame(&ops)
		now = now.Add(10 * time.Millisecond)
	}
	click := func(key string) {
		u.btn(key).Click()
		frame()
		frame()
	}
	frame()
	return u, b, frame, click
}

func visibleIDs(u *UI) []string {
	var ids []string
	for _, c := range u.sidebar.visible {
		ids = append(ids, c.ID)
	}
	return ids
}

// TestListChips checks that each custom list has a chip that shows its
// chats, and that the "+" chip makes a new list.
func TestListChips(t *testing.T) {
	u, b, frame, click := chatListUI(t)
	items := u.chipItems()
	var names []string
	for _, it := range items {
		names = append(names, it.name)
	}
	if want := []string{"All", "Unread", "Favourites", "Groups", "Family", "Work", "New list"}; !slices.Equal(names, want) {
		t.Fatalf("chips %q, want %q", names, want)
	}
	click("chip:4") // Family
	if got := visibleIDs(u); !slices.Equal(got, []string{"fams", "ibuk"}) && !slices.Equal(got, []string{"ibuk", "fams"}) {
		t.Errorf("Family shows %v", got)
	}

	click("chip:6") // +
	if u.dialog.kind != dialogForward || !u.dialog.newList {
		t.Fatal("the + chip didn't open New list")
	}
	u.dialog.listName.SetText("Pelatnas")
	click("fwd:pelatnas")
	frame()
	click("fwd:send")
	for range 30 {
		frame()
	}
	ls := b.Lists()
	if l := ls[len(ls)-1]; l.Name != "Pelatnas" || !slices.Equal(l.Chats, []string{"pelatnas"}) {
		t.Fatalf("new list %q with %v", l.Name, l.Chats)
	}
	if n := len(u.chipItems()); n != 8 {
		t.Errorf("%d chips after the new list, want 8", n)
	}
}

// TestListSearch checks that the chat list's search finds archived chats,
// contacts you have no chat with and messages.
func TestListSearch(t *testing.T) {
	u, b, frame, _ := chatListUI(t)
	search := func(q string) []findRow {
		u.sidebar.search.SetText(q)
		for range 3 {
			frame()
		}
		return u.sidebar.find.rows
	}
	rows := search("old project")
	if len(rows) < 2 || rows[0].heading != "Chats" || rows[1].chat == nil || rows[1].chat.ID != "old-project" {
		t.Errorf("an archived chat isn't found: %+v", rows)
	}

	var contact *model.Contact
	for _, c := range b.Contacts() {
		if u.chatByID(c.ID) == nil {
			contact = c
			break
		}
	}
	if contact == nil {
		t.Fatal("the demo has no contact without a chat")
	}
	found := false
	for _, r := range search(contact.Name) {
		found = found || r.contact != nil && r.contact.ID == contact.ID
	}
	if !found {
		t.Errorf("contact %s isn't found", contact.Name)
	}

	var heading bool
	var msgs int
	for _, r := range search("staging freeze") {
		heading = heading || r.heading == "Messages"
		if r.msg != nil {
			msgs++
		}
	}
	if !heading || msgs == 0 {
		t.Errorf("messages aren't found (heading %v, %d messages)", heading, msgs)
	}
}

// TestChatRowMenu checks the chat row's menu follows WhatsApp's: Block for
// a contact, and no Close chat even for the open chat.
func TestChatRowMenu(t *testing.T) {
	u, _, _, _ := chatListUI(t)
	u.SelectID("citra")
	var labels []string
	for _, it := range u.chatMenuItems(u.chatByID("citra")) {
		labels = append(labels, it.label)
	}
	want := []string{"Archive chat", "Unmute notifications", "Unpin chat", "Mark as unread", "Add to favourites",
		"Add to list", "", "Block", "Clear chat", "Delete chat"}
	// The demo's Citra may be muted or not, favourite or not, unread or not:
	// compare the order of what's there.
	var got []string
	for _, l := range labels {
		switch l {
		case "Mute notifications":
			l = "Unmute notifications"
		case "Pin chat":
			l = "Unpin chat"
		case "Mark as read":
			l = "Mark as unread"
		case "Remove from favourites":
			l = "Add to favourites"
		}
		got = append(got, l)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Citra's menu is %q, want %q", labels, want)
	}
	for _, it := range u.chatMenuItems(u.chatByID("fams")) {
		if it.label == "Block" || it.label == "Close chat" {
			t.Errorf("a group's menu offers %q", it.label)
		}
	}
}

// TestSelectChats picks chats and runs the select bar's actions on them.
func TestSelectChats(t *testing.T) {
	u, _, frame, _ := chatListUI(t)
	u.startChatSelect()
	frame()
	u.pickRow("arman")
	u.pickRow("lintang")
	u.pickAction("pin")
	frame()
	if u.sidebar.pick.on {
		t.Error("select mode is still on after Pin")
	}
	for _, id := range []string{"arman", "lintang"} {
		if !u.chatByID(id).Pinned {
			t.Errorf("%s isn't pinned", id)
		}
	}

	u.startChatSelect()
	u.pickRow("arman")
	u.pickRow("lintang")
	u.pickAction("archive")
	frame()
	for _, id := range []string{"arman", "lintang"} {
		if !u.chatByID(id).Archived {
			t.Errorf("%s isn't archived", id)
		}
	}

	// Mute asks how long first; mark read takes the unread chats.
	u.startChatSelect()
	u.pickRow("citra")
	u.pickRow("work")
	u.pickAction("mute")
	if u.ctx.kind != ctxPickMute || !u.sidebar.pick.on {
		t.Fatalf("Mute opened menu %v, select mode %v", u.ctx.kind, u.sidebar.pick.on)
	}
	u.pickMuteItems()[2].run() // Mute always
	frame()
	if !u.chatByID("citra").Muted || !u.chatByID("work").Muted {
		t.Error("Mute always didn't mute both chats")
	}
	u.startChatSelect()
	u.pickRow("work")
	if u.chatByID("work").Unread == 0 {
		t.Fatal("the demo's work chat isn't unread")
	}
	u.pickAction("read")
	frame()
	if n := u.chatByID("work").Unread; n != 0 {
		t.Errorf("work still has %d unread", n)
	}

	// Delete leaves groups alone.
	u.startChatSelect()
	u.pickRow("bank")
	u.pickRow("design")
	u.pickAction("delete")
	if u.dialog.kind != dialogConfirm {
		t.Fatal("Delete didn't ask")
	}
	u.dialog.buttons[0].run()
	frame()
	if u.chatByID("bank") != nil {
		t.Error("bank wasn't deleted")
	}
	if u.chatByID("design") == nil {
		t.Error("a group was deleted")
	}

	// Escape leaves select mode.
	u.dialog = dialogState{}
	u.startChatSelect()
	u.Escape()
	if u.sidebar.pick.on {
		t.Error("Escape didn't leave select mode")
	}
}

// TestSelectRange checks the shift-click range of the select mode: it
// picks the chats from the last plain-clicked row to the clicked one, and
// grows or clears with more shift-clicks.
func TestSelectRange(t *testing.T) {
	u, _, frame, _ := chatListUI(t)
	u.startChatSelect()
	frame()
	u.pickRow("arman")
	u.pickRowAt("bank", true) // range arman..bank, in u.chats order
	between := map[string]bool{}
	for _, c := range rangeBetween(u, "arman", "bank") {
		between[c.ID] = true
	}
	got := map[string]bool{}
	for _, c := range u.pickedChats() {
		got[c.ID] = true
	}
	if !slices.Equal(keysOf(got), keysOf(between)) {
		t.Errorf("shift-click picked %d chats, want the %d between arman and bank", len(got), len(between))
	}
	if u.sidebar.pick.anchor != "arman" {
		t.Error("a shift-click moved the anchor")
	}

	// A shift-click on a picked row clears its range instead: what stays
	// is what was picked beyond the cleared range.
	u.pickRowAt("pelatnas", true)
	cleared := map[string]bool{}
	for _, c := range rangeBetween(u, "arman", "pelatnas") {
		cleared[c.ID] = true
	}
	var kept []*model.Chat
	for _, c := range u.pickedChats() {
		if !cleared[c.ID] {
			kept = append(kept, c)
		}
	}
	wantKept := len(rangeBetween(u, "pelatnas", "bank")) - 1 // bank's span past pelatnas
	if len(kept) != wantKept {
		t.Errorf("clearing kept %d chats (%v), want %d", len(kept), keptIDs(kept), wantKept)
	}
	u.pickRowAt("citra", true) // grows: anchor still arman
	if n := len(u.pickedChats()); n <= len(kept) {
		t.Errorf("the range didn't grow back (%d chats)", n)
	}

	// A shift-click without an anchor picks just the one chat.
	u.endChatSelect()
	u.startChatSelect()
	u.pickRowAt("arman", true)
	if n := len(u.pickedChats()); n != 1 || !u.sidebar.pick.picked["arman"] {
		t.Errorf("a lone shift-click picked %d chats", n)
	}
}

// TestSelectMsgRange checks the shift-click range of the conversation's
// select mode: it picks the messages from the last plain-clicked row to
// the clicked one, past the system notes, and clears with more
// shift-clicks.
func TestSelectMsgRange(t *testing.T) {
	u, _, frame, _ := chatListUI(t)
	u.SelectID("citra")
	frame()
	u.startSelectNone()
	frame()
	n := len(u.msgs)
	if n < 3 {
		t.Fatalf("the demo chat has %d messages", n)
	}
	first, last := u.msgs[0], u.msgs[n-1]
	u.pickMsg([]*model.Message{first}, false)
	u.pickMsg([]*model.Message{last}, true)
	want := 0
	between := false
	for _, m := range u.msgs {
		if m.ID == first.ID {
			between = true
		}
		if between && m.Kind != model.KindSystem {
			want++
		}
	}
	if got := len(u.pickedMessages()); got != want {
		t.Errorf("shift-click picked %d messages, want the %d from first to last", got, want)
	}
	if u.conv.selAnchor != first.ID {
		t.Error("a shift-click moved the anchor")
	}

	// A shift-click on a picked row clears its range: what stays is the
	// span past the cleared one.
	mid := u.msgs[1]
	u.pickMsg([]*model.Message{mid}, true)
	wantKept := 0
	for i := 2; i < n; i++ {
		if u.msgs[i].Kind != model.KindSystem {
			wantKept++
		}
	}
	if got := len(u.pickedMessages()); got != wantKept {
		t.Errorf("clearing kept %d messages, want %d", got, wantKept)
	}

	// A shift-click without an anchor picks just the one message.
	u.endSelect()
	u.startSelectNone()
	u.pickMsg([]*model.Message{last}, true)
	if got := len(u.pickedMessages()); got != 1 || !u.conv.picked[last.ID] {
		t.Errorf("a lone shift-click picked %d messages", got)
	}
}

// rangeBetween returns the chats from one to the other, in the list's
// order, the range a shift-click picks.
func rangeBetween(u *UI, a, b string) []*model.Chat {
	picked := false
	var out []*model.Chat
	for _, c := range u.chats {
		if c.ID == a || c.ID == b {
			picked = !picked
		}
		if picked {
			out = append(out, c)
		}
	}
	if c := u.chatByID(b); c != nil && (len(out) == 0 || out[len(out)-1].ID != b) {
		out = append(out, c)
	}
	return out
}

// keptIDs lists a slice of chats' IDs, for error messages.
func keptIDs(cs []*model.Chat) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.ID)
	}
	return out
}

// keysOf returns a shallow copy of a set, so two maps can be compared with
// slices.Equal regardless of ordering.
func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// TestAppLock sets a password, locks, unlocks with it and turns it off.
func TestAppLock(t *testing.T) {
	stored := hashPassword("secret1")
	if !checkPassword(stored, "secret1") || checkPassword(stored, "secret2") || checkPassword("", "") {
		t.Fatal("checkPassword got it wrong")
	}
	if hashPassword("secret1") == stored {
		t.Error("two hashes of a password are the same: no salt")
	}

	u, _, frame, _ := chatListUI(t)
	if u.lock.locked {
		t.Fatal("locked without a password")
	}
	u.backend.SetPref(prefAppLock, stored)
	u.backend.SetPref(prefAppLockAfter, "1")
	u.touchLock()
	frame()
	if u.lock.locked {
		t.Fatal("locked at once")
	}
	// A minute without input locks it.
	u.now = func() time.Time { return testNow().Add(2 * time.Minute) }
	frame()
	if !u.lock.locked {
		t.Fatal("didn't lock itself after a minute")
	}
	u.lock.pw.SetText("wrong!")
	u.btn("lock:unlock").Click()
	frame()
	if !u.lock.locked || !u.lock.wrong {
		t.Fatal("a wrong password unlocked it")
	}
	u.lock.pw.SetText("secret1")
	u.btn("lock:unlock").Click()
	frame()
	if u.lock.locked {
		t.Fatal("the password didn't unlock it")
	}
	// A new UI (a new window) starts locked.
	if v := New(u.backend); !v.lock.locked {
		t.Error("a new UI isn't locked")
	}
	// Turned off, the lock goes.
	u.lockNow()
	u.backend.SetPref(prefAppLock, "")
	frame()
	if u.lock.locked {
		t.Error("still locked with App lock off")
	}
}
