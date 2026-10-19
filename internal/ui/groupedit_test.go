package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"slices"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

type infoHarness struct {
	t     *testing.T
	b     *mock.Backend
	u     *UI
	r     input.Router
	ops   op.Ops
	now   time.Time
	frame func()
}

func newInfoHarness(t *testing.T, chat string) *infoHarness {
	h := &infoHarness{t: t, b: mock.New(), now: testNow()}
	h.u = New(h.b)
	h.u.Start(func() {})
	h.u.SelectID(chat)
	h.frame = func() {
		h.ops.Reset()
		h.u.Layout(layout.Context{Ops: &h.ops, Now: h.now, Source: h.r.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1400, 900))})
		h.r.Frame(&h.ops)
		h.now = h.now.Add(10 * time.Millisecond)
	}
	h.u.openInfo(chat)
	h.frames(30)
	return h
}

func (h *infoHarness) frames(n int) {
	for range n {
		h.frame()
	}
}

func (h *infoHarness) click(key string) {
	h.u.btn(key).Click()
	h.frames(2)
}

func (h *infoHarness) press(name key.Name) {
	h.r.Queue(key.Event{Name: name, State: key.Press}, key.Event{Name: name, State: key.Release})
	h.frames(2)
}

// TestEditGroupInfo: the name and description pencils edit them in place,
// Enter or the tick saves, Escape cancels, and the panel shows what was
// saved at once.
func TestEditGroupInfo(t *testing.T) {
	h := newInfoHarness(t, "fams")
	u, b := h.u, h.b
	e := &u.info.edit

	h.click("info:editname")
	if e.field != editGroupName {
		t.Fatal("the name's pencil didn't start editing")
	}
	if got := e.editor.Text(); got != b.Info("fams").Name {
		t.Errorf("the editor holds %q", got)
	}
	e.editor.SetText("  Family 🫶 ")
	e.save.Click()
	h.frames(3)
	if e.field != 0 {
		t.Fatal("the tick didn't save")
	}
	if got := b.Info("fams").Name; got != "Family 🫶" {
		t.Errorf("the group is named %q", got)
	}
	if c := u.chatByID("fams"); c.Name != "Family 🫶" {
		t.Errorf("the chat list says %q", c.Name)
	}

	// An empty name isn't saved.
	h.click("info:editname")
	e.editor.SetText(" ")
	e.save.Click()
	h.frames(2)
	if e.field != editGroupName || b.Info("fams").Name != "Family 🫶" {
		t.Error("an empty name was saved")
	}
	h.press(key.NameEscape)
	if e.field != 0 {
		t.Fatal("Escape didn't stop editing")
	}

	h.click("info:editdesc")
	if e.field != editGroupDesc || e.editor.SingleLine {
		t.Fatal("the description's pencil didn't start a multi-line editor")
	}
	h.frames(2) // focus
	e.editor.SetText("First line\nsecond line")
	h.press(key.NameReturn)
	if e.field != 0 {
		t.Fatal("Enter didn't save the description")
	}
	if got := b.Info("fams").About; got != "First line\nsecond line" {
		t.Errorf("the description is %q", got)
	}

	// What's saved shows before the group's info has it.
	u.info.edit.pending = map[pendingKey]pendingValue{{"fams", editGroupDesc}: {"On its way", u.now()}}
	if got := u.withPending("fams", b.Info("fams")).About; got != "On its way" {
		t.Errorf("the panel shows %q while the change is on its way", got)
	}
	if b.Info("fams").About == "On its way" {
		t.Error("withPending changed the backend's info")
	}
	b.Info("fams").About = "On its way"
	u.withPending("fams", b.Info("fams"))
	if len(u.info.edit.pending) != 0 {
		t.Error("the change was kept after it arrived")
	}
}

// TestEditGroupLocked: when only admins may edit the group's info, a
// member sees no pencils and is told why.
func TestEditGroupLocked(t *testing.T) {
	h := newInfoHarness(t, "fams")
	u, b := h.u, h.b
	// The mock keeps a group's info once it changed.
	b.ManageGroup(model.GroupRequest{ChatID: "fams", Action: model.GroupLock, On: true})
	info := b.Info("fams")
	for i := range info.Members {
		if info.Members[i].Me {
			info.Members[i].Admin = false
		}
	}
	h.frames(2)
	c := u.chatByID("fams")
	if u.canEditGroup(c, info) {
		t.Fatal("a member may edit a locked group")
	}
	u.startGroupEdit(c, info, editGroupName)
	if u.info.edit.field != 0 || !u.dialog.isOpen() {
		t.Error("editing a locked group didn't explain why not")
	}
	u.closeDialog()
	h.frames(30)
	h.click("info:photo")
	if u.ctx.isOpen() {
		t.Error("a member got the group photo's menu")
	}
}

func testJPEG(t *testing.T, c color.NRGBA) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestGroupPhoto: the group's picture opens a menu that views, changes
// and removes it.
func TestGroupPhoto(t *testing.T) {
	h := newInfoHarness(t, "fams")
	u, b := h.u, h.b
	pic := testJPEG(t, color.NRGBA{R: 200, A: 255})
	u.setGroupPhoto("fams", pic)
	h.frames(5)
	if !bytes.Equal(b.Avatar("fams"), pic) {
		t.Fatal("the new picture wasn't set")
	}
	waitFor(t, h, func() bool { return u.avatarImage("fams") != nil })

	h.click("info:photo")
	if u.ctx.kind != ctxInfoPhoto || !u.ctx.isOpen() {
		t.Fatalf("the picture opened menu %v", u.ctx.kind)
	}
	var keys []string
	for _, it := range u.infoPhotoItems("fams") {
		keys = append(keys, it.key)
	}
	if want := []string{"iphoto:view", "iphoto:change", "iphoto:remove"}; !slices.Equal(keys, want) {
		t.Errorf("the menu has %q, want %q", keys, want)
	}
	h.click("menu:iphoto:view")
	if !u.picView.isOpen() || u.picView.id != "fams" {
		t.Fatal("View photo didn't show it")
	}
	waitFor(t, h, func() bool {
		u.images.mu.Lock()
		defer u.images.mu.Unlock()
		e := u.images.m["pv:fams"]
		return e != nil && e.state == imgReady
	})
	h.press(key.NameEscape)
	h.frames(40)
	if u.picView.isOpen() || u.picView.anim.v != 0 {
		t.Error("Escape didn't close the picture")
	}

	u.setGroupPhoto("fams", nil)
	h.frames(5)
	waitFor(t, h, func() bool { return b.Avatar("fams") == nil && u.avatarImage("fams") == nil })
	var labels []string
	for _, it := range u.infoPhotoItems("fams") {
		labels = append(labels, it.label)
	}
	if !slices.Equal(labels, []string{"Upload photo"}) {
		t.Errorf("without a picture the menu has %q", labels)
	}
}

// TestContactPhoto: a person's picture in their info shows it big.
func TestContactPhoto(t *testing.T) {
	h := newInfoHarness(t, "arman")
	u := h.u
	if u.backend.Avatar("arman") == nil {
		t.Skip("the demo contact has no picture")
	}
	waitFor(t, h, func() bool { return u.avatarImage("arman") != nil })
	h.click("info:photo")
	if !u.picView.isOpen() || u.picView.id != "arman" || u.ctx.isOpen() {
		t.Fatal("clicking a contact's picture didn't show it")
	}
	h.click("pv:close")
	if u.picView.isOpen() {
		t.Error("the close button didn't close it")
	}
}

func waitFor(t *testing.T, h *infoHarness, ok func() bool) {
	t.Helper()
	for end := time.Now().Add(5 * time.Second); !ok(); time.Sleep(5 * time.Millisecond) {
		if time.Now().After(end) {
			t.Fatal("timed out")
		}
		h.frame()
	}
}
