package ui

import (
	"slices"
	"testing"
	"time"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// TestConvMenu checks the chat's ⋮ menu offers a group's and a contact's
// actions, and that its items do something.
func TestConvMenu(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	labels := func(id string) []string {
		u.SelectID(id)
		var out []string
		for _, it := range u.convMenuItems(u.selected) {
			if it.run == nil && !it.divider {
				t.Errorf("%s: %q does nothing", id, it.label)
			}
			out = append(out, it.label)
		}
		return out
	}
	group := labels("fams")
	for _, want := range []string{"Add member", "Group info", "Search", "Select messages", "Disappearing messages",
		"Chat theme", "Add to list", "Export chat", "Close chat", "Clear chat", "Exit group"} {
		if !slices.Contains(group, want) {
			t.Errorf("group menu %q lacks %q", group, want)
		}
	}
	contact := labels("citra")
	for _, want := range []string{"Contact info", "Block", "Report", "Delete chat"} {
		if !slices.Contains(contact, want) {
			t.Errorf("contact menu %q lacks %q", contact, want)
		}
	}
	if slices.Contains(contact, "Add member") || slices.Contains(contact, "Exit group") {
		t.Errorf("contact menu %q has group actions", contact)
	}
}

// TestDisappearingAndTheme sets a chat's timer and theme.
func TestDisappearingAndTheme(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.SelectID("fams")
	var week *menuItem
	items := u.timerItems(u.selected)
	for i := range items {
		if items[i].label == "7 days" {
			week = &items[i]
		}
	}
	if week == nil {
		t.Fatal("no 7 days timer")
	}
	week.run()
	u.applyEvents()
	if info := u.backend.Info("fams"); info.Disappearing != uint32(7*24*time.Hour/time.Second) {
		t.Errorf("timer = %d s, want 7 days", info.Disappearing)
	}
	u.setChatTheme("fams", 3)
	u.themes = nil // read back from the prefs
	if got := u.chatThemeOf("fams"); got != 3 {
		t.Errorf("theme = %d, want 3", got)
	}
	if u.chatPalette(u.selected).BubbleOut == u.pal.BubbleOut {
		t.Error("the theme didn't change your bubbles")
	}
}

// TestOldChatTheme reads a theme kept before WhatsApp's themes as the one
// it became.
func TestOldChatTheme(t *testing.T) {
	u := New(mock.New())
	u.backend.SetPref(prefChatTheme+"fams", "Teal")
	if got := chatThemes[u.chatThemeOf("fams")].id; got != "Emerald" {
		t.Errorf("Teal became %q, want Emerald", got)
	}
}

// TestChatLook sets the look of every chat, then one chat's bubbles and
// wallpaper apart from it.
func TestChatLook(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.SetDark(true)
	pink := schemeIndex("Pink")
	u.setChatTheme(lookAll, pink)
	if got := u.chatThemeOf("fams"); got != pink {
		t.Fatalf("fams's scheme = %d, want every chat's %d", got, pink)
	}
	// Its bubbles alone: the wallpaper stays Pink's.
	red := schemeIndex("Red@Tonal")
	u.setBubbles("fams", red)
	l := u.lookOf("fams")
	if l.scheme != red || l.wall != "theme:Pink" {
		t.Errorf("look = %+v, want Red@Tonal on Pink's wallpaper", l)
	}
	if p := u.lookPalette(l); p.BubbleOut != rgb(chatThemes[red].dark.out) || p.ChatBg != rgb(chatThemes[pink].dark.wall) {
		t.Errorf("palette has bubbles %v on %v", p.BubbleOut, p.ChatBg)
	}
	// A plain color, read back from the prefs.
	u.setWall("fams", "color:3")
	u.themes = nil
	if p := u.lookPalette(u.lookOf("fams")); p.ChatBg != solidLook(solidWalls[3], true).bg {
		t.Errorf("wallpaper = %v, want %s", p.ChatBg, solidWalls[3].name)
	}
	// Default for one chat while every chat has Pink.
	u.setChatTheme("fams", 0)
	if got := u.chatThemeOf("fams"); got != 0 {
		t.Errorf("fams's scheme = %d, want the app's own", got)
	}
	if got := u.chatThemeOf("citra"); got != pink {
		t.Errorf("citra's scheme = %d, want every chat's %d", got, pink)
	}
}

// TestAppAccent picks an App theme color and goes back to WhatsApp's green.
func TestAppAccent(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.SetDark(true)
	u.setAccent("Red")
	if u.pal.Green != rgb(0xef3345) || u.pal.ChipActive != rgb(0xa71136) {
		t.Errorf("dark Red accent = %v, chip %v", u.pal.Green, u.pal.ChipActive)
	}
	u.SetDark(false)
	if u.pal.Green != rgb(0xea0038) {
		t.Errorf("light Red accent = %v", u.pal.Green)
	}
	if u.backend.Pref(prefAccent) != "Red" {
		t.Errorf("pref = %q", u.backend.Pref(prefAccent))
	}
	u.setAccent("WhatsApp-Green")
	if u.pal != &lightPalette {
		t.Error("WhatsApp's green isn't the plain palette")
	}
}

// TestGalleryPanel opens the Media panel and switches its tabs.
func TestGalleryPanel(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.openGallery("", "")
	u.applyEvents()
	if len(u.gallery.list.msgs) == 0 {
		t.Fatal("no media from all chats")
	}
	for _, m := range u.gallery.list.msgs {
		if m.Kind != model.KindImage && m.Media != model.MediaImage && m.Media != model.MediaVideo {
			t.Errorf("media tab lists %+v", m)
		}
	}
	u.gallery.tab = model.GalleryDocs
	u.loadGallery()
	u.applyEvents()
	if len(u.gallery.list.msgs) == 0 {
		t.Fatal("no documents")
	}
	for _, m := range u.gallery.list.msgs {
		if m.Media != model.MediaDocument {
			t.Errorf("docs tab lists %+v", m)
		}
	}
	u.closeGallery()
	if u.gallery.open {
		t.Error("the panel stayed open")
	}
}

// TestChatThemeReset opens a chat's Chat theme panel, changes its look and
// resets it to every chat's.
func TestChatThemeReset(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.applyEvents()
	u.SelectID("fams")
	pink := schemeIndex("Pink")
	u.setChatTheme(lookAll, pink)
	u.openChatTheme(u.selected)
	if !u.themeUI.open || u.dialog.isOpen() {
		t.Fatal("a chat's theme should open as a panel")
	}
	u.setChatTheme("fams", 0) // Default, apart from every chat's
	l := u.lookOf("fams")
	l.bare = true
	u.setLook("fams", l)
	u.themes = nil
	if got := u.lookOf("fams"); got != (chatLook{bare: true}) {
		t.Fatalf("look = %+v, want Default without doodles", got)
	}
	u.resetLook("fams")
	if got := u.lookOf("fams"); got != (chatLook{scheme: pink}) {
		t.Errorf("after the reset look = %+v, want every chat's Pink", got)
	}
	u.escape()
	if u.themeUI.open || u.selected == nil {
		t.Error("Esc should close the panel alone")
	}
}
