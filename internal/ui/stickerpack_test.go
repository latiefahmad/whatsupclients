package ui

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// packMessage returns the first demo message that shares a sticker pack.
func packMessage(t *testing.T, st *slashTest) *model.Message {
	t.Helper()
	for _, m := range st.u.msgs {
		if m.Pack != nil {
			return m
		}
	}
	t.Fatal("the chat shares no sticker pack")
	return nil
}

// A shared pack's card has a "View sticker pack" button, which opens the
// dialog and asks the backend for the pack.
func TestStickerPackCard(t *testing.T) {
	st := newSlashTest(t, "ririiiin")
	m := packMessage(t, st)
	bs := cardButtons(m)
	if len(bs) != 1 || bs[0].Kind != buttonPack || bs[0].Label != "View sticker pack" {
		t.Fatalf("card buttons = %+v", bs)
	}
	st.u.pressCardButton(m, bs[0])
	st.u.applyEvents()
	if st.u.dialog.kind != dialogPack || st.u.dialog.pack.pack == nil {
		t.Fatalf("dialog = %v, pack %v", st.u.dialog.kind, st.u.dialog.pack.pack)
	}
	if len(st.u.dialog.pack.stickers) == 0 {
		t.Fatal("the pack's stickers never came")
	}
	st.frame()
}

// A pack added to the picker becomes a tab of its own, which lists its
// stickers.
func TestStickerPackTab(t *testing.T) {
	st := newSlashTest(t, "ririiiin")
	m := packMessage(t, st)
	st.u.openPackView(m)
	st.u.applyEvents()
	st.u.backend.AddStickerPack(m.Pack.ID, true)
	st.u.applyEvents()
	st.u.picker.packsOK = false // as a StickersEvent would

	packs := st.u.pickerPacks()
	if len(packs) != 1 || packs[0].ID != m.Pack.ID {
		t.Fatalf("picker packs = %+v", packs)
	}
	set := model.StickerSet(len(stickerSets))
	got := st.u.stickerList(set)
	if len(got) == 0 {
		t.Fatal("the tab lists no stickers")
	}
	// Taking it off leaves the picker on a set that exists.
	st.u.backend.AddStickerPack(m.Pack.ID, false)
	st.u.applyEvents()
	st.u.picker.packsOK = false
	st.u.picker.stickerSet = set
	st.u.picker.composerTab = tabSticker
	st.u.openPicker(pickComposer, nil)
	st.frame() // layoutStickerTab falls back to a set that is there
	if int(st.u.picker.stickerSet) >= len(stickerSets) {
		t.Fatalf("stuck on the pack that was taken off: %v", st.u.picker.stickerSet)
	}
}

// The "pack" overlay draws the pack's dialog, and "packtab" the picker with
// the pack's tab: both lay a frame out without panicking.
func TestPackOverlaysDraw(t *testing.T) {
	for _, name := range []string{"pack", "packtab"} {
		b := mock.New()
		u := New(b)
		u.Start(func() {})
		u.applyEvents()
		u.SelectID("ririiiin")
		u.ShowOverlay(name, 0, 0)
		u.applyEvents()
		var ops op.Ops
		var r input.Router
		u.Layout(layout.Context{Ops: &ops, Now: testNow(), Source: r.Source(),
			Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Exact(image.Pt(1100, 700))})
		r.Frame(&ops)
	}
}
