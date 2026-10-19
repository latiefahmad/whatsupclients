package ui

import (
	"image"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// TestLinkSpans checks which part of a text is a link: not the punctuation
// that ends a sentence, unless it belongs to the link.
func TestLinkSpans(t *testing.T) {
	u := New(mock.New())
	for _, tc := range []struct{ text, link string }{
		{"see https://example.com.", "https://example.com"},
		{"(www.example.com/a)", "www.example.com/a"},
		{"https://en.wikipedia.org/wiki/Go_(game), nice", "https://en.wikipedia.org/wiki/Go_(game)"},
		{"is it https://example.com/?q=1?", "https://example.com/?q=1"},
	} {
		spans, deco := u.richSpans(tc.text, 15, u.pal.Text, false, 0)
		got := ""
		for i, s := range spans {
			if deco[i]&decoLink != 0 {
				got += s.Content
			}
		}
		if got != tc.link {
			t.Errorf("%q: link %q, want %q", tc.text, got, tc.link)
		}
	}
}

func TestInviteCode(t *testing.T) {
	for link, want := range map[string]string{
		"https://chat.whatsapp.com/FrLhsO1aBMEBhOFBtpLwxl":        "FrLhsO1aBMEBhOFBtpLwxl",
		"https://chat.whatsapp.com/invite/FrLhsO1aBMEBhOFBtpLwxl": "FrLhsO1aBMEBhOFBtpLwxl",
		"http://CHAT.whatsapp.com/abc123/":                        "abc123",
		"https://chat.whatsapp.com/":                              "",
		"https://chat.whatsapp.com.evil.example/abc":              "",
		"https://example.com/abc":                                 "",
		"www.chat.whatsapp.com/abc":                               "",
	} {
		if got := inviteCode(link); got != want {
			t.Errorf("inviteCode(%q) = %q, want %q", link, got, want)
		}
	}
}

// TestLinkHoverClick checks that hovering a link takes its underline away
// and that clicking an invite link opens the invite dialog.
func TestLinkHoverClick(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.applyEvents()
	now := testNow()
	var ops op.Ops
	var r input.Router
	frame := func() {
		ops.Reset()
		gtx := layout.Context{Ops: &ops, Now: now, Source: r.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Constraints{Max: image.Pt(600, 400)}}
		u.layoutRich(gtx, "https://chat.whatsapp.com/DemoInviteFutsal is the link", 15, u.pal.Text, u.pal.TextSecondary,
			richOpts{links: "t"})
		r.Frame(&ops)
		now = now.Add(10 * time.Millisecond)
	}
	frame()
	key := "lk:t:0:0"
	if u.btn(key).Hovered() {
		t.Fatal("the link is hovered before the pointer moved")
	}
	p := f32.Pt(30, 10)
	r.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p})
	frame()
	if !u.btn(key).Hovered() {
		t.Fatal("the link isn't hovered under the pointer")
	}
	r.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: p, Buttons: pointer.ButtonPrimary},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
	frame()
	frame()
	if u.dialog.kind != dialogInvite || u.dialog.invite.code != "DemoInviteFutsal" {
		t.Fatalf("clicking the link opened dialog %d (code %q), want the invite", u.dialog.kind, u.dialog.invite.code)
	}
}

// TestOpenAppLink checks that a wa.me link opens the number's chat with
// the link's text in the composer.
func TestOpenAppLink(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.applyEvents()
	u.openOutsideLink("whatsapp://send?phone=%2B1+555-123-4567&text=Hello%21")
	u.applyEvents()
	if u.selected == nil || u.selected.ID != "15551234567@s.whatsapp.net" {
		t.Fatalf("open chat %v, want 15551234567@s.whatsapp.net", u.selected)
	}
	if got := u.conv.composer.Text(); got != "Hello!" {
		t.Errorf("composer %q, want %q", got, "Hello!")
	}
}

// TestResolveLinks opens the links the server resolves with the demo
// backend: a business's message link and a contact's QR link open their
// chats, a followed channel opens, and another shows to follow first.
func TestResolveLinks(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.applyEvents()
	open := func(link string) {
		t.Helper()
		if !u.openAppLink(link) {
			t.Fatalf("openAppLink(%q) = false", link)
		}
		u.applyEvents()
	}
	open("https://wa.me/message/" + mock.LinkBusiness)
	if u.selected == nil || u.selected.ID != "vivy@lid" {
		t.Fatalf("business link opened %v, want vivy@lid", u.selected)
	}
	if got := u.conv.composer.Text(); !strings.HasPrefix(got, "Hi! I'd like") {
		t.Errorf("business link's composer %q", got)
	}
	open("https://wa.me/qr/" + mock.LinkContact)
	if u.selected == nil || u.selected.ID != "arman" {
		t.Fatalf("contact link opened %v, want arman", u.selected)
	}
	open("https://whatsapp.com/channel/" + mock.LinkChannelFollowed)
	if u.page != pageChannels || u.selected == nil || u.selected.ID != "techdaily@newsletter" {
		t.Fatalf("followed channel's link: page %v, chat %v", u.page, u.selected)
	}
	open("https://whatsapp.com/channel/" + mock.LinkChannelSuggested)
	in := u.dialog.invite
	if u.dialog.kind != dialogInvite || !in.channel || in.ch == nil || in.ch.ID != "bola@newsletter" {
		t.Fatalf("suggested channel's link shows %+v", in)
	}
	u.followFromLink(in.ch.ID)
	u.applyEvents()
	if u.selected == nil || u.selected.ID != "bola@newsletter" {
		t.Fatalf("followed channel opened %v, want bola@newsletter", u.selected)
	}
	open("https://wa.me/qr/Unknown1")
	if u.selected.ID != "bola@newsletter" {
		t.Errorf("an unknown link opened %v", u.selected)
	}
}

// TestShareLink checks that a share link (text, no number) asks for the
// chat, and that the chat picked opens with the text in its composer.
func TestShareLink(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.applyEvents()
	const link = "https://api.whatsapp.com/send/?text=https%3A%2F%2Fwww.reddit.com%2Fr%2Fx&type=custom_url&app_absent=0"
	if !u.openAppLink(link) {
		t.Fatal("the share link wasn't taken")
	}
	if u.dialog.kind != dialogForward || u.dialog.shareText != "https://www.reddit.com/r/x" {
		t.Fatalf("dialog %v with %q", u.dialog.kind, u.dialog.shareText)
	}
	var ops op.Ops
	var r input.Router
	frame := func() {
		ops.Reset()
		u.Layout(layout.Context{Ops: &ops, Now: testNow(), Source: r.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1100, 700))})
		r.Frame(&ops)
	}
	frame()
	u.btn("fwd:citra").Click()
	frame()
	if u.selected == nil || u.selected.ID != "citra" {
		t.Fatalf("picked citra, open %v", u.selected)
	}
	if got := u.conv.composer.Text(); got != "https://www.reddit.com/r/x" {
		t.Errorf("composer %q", got)
	}
	if u.dialog.isOpen() {
		t.Error("the picker is still open")
	}
}

// TestMentionClick checks that only mentions of people are clickable, and
// that clicking one shows their contact info.
func TestMentionClick(t *testing.T) {
	b := mock.New()
	u := New(b)
	u.Start(func() {})
	u.applyEvents()
	var m *model.Message
	for _, x := range b.Messages("work", 100) {
		if strings.Contains(x.Text, "@Mas Rif") {
			m = x
		}
	}
	if m == nil {
		t.Fatal("no demo message mentions Mas Rif")
	}
	spans, deco := u.richSpans(m.Text, 15, u.pal.Text, false, pillMe)
	got := ""
	for i, s := range spans {
		if deco[i]&decoMention != 0 {
			got += s.Content
		}
	}
	if got != "@Mas Rif" {
		t.Fatalf("clickable mentions %q, want only @Mas Rif", got)
	}
	u.openChatByID("work")
	u.openMention(m, "@Mas Rif")
	if !u.info.open || u.info.chatID != "mas rif" || u.selected.ID != "work" {
		t.Fatalf("clicking @Mas Rif showed info %q (open %v) in chat %q, want mas rif's in work", u.info.chatID, u.info.open, u.selected.ID)
	}
}
