package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

func TestParseChatLink(t *testing.T) {
	for link, want := range map[string]*appLink{
		"whatsup://send?phone=6281234567890":                                                       {phone: "6281234567890"},
		"whatsup://send?phone=6281234567890&text=hi%20there":                                       {phone: "6281234567890", text: "hi there"},
		"whatsup://6281234567890?text=hello":                                                       {phone: "6281234567890", text: "hello"},
		"whatsup:6281234567890":                                                                    {phone: "6281234567890"},
		`"https://wa.me/6281234567890"`:                                                            {phone: "6281234567890"},
		"https://wa.me/6281234567890":                                                              {phone: "6281234567890"},
		"https://wa.me/6281234567890?text=order%20status%3F":                                       {phone: "6281234567890", text: "order status?"},
		"http://wa.me/6281234567890":                                                               {phone: "6281234567890"},
		"https://api.whatsapp.com/send?phone=6281234567890&text=hi":                                {phone: "6281234567890", text: "hi"},
		"whatsapp://send?phone=6281234567890&text=Hi%20there":                                      {phone: "6281234567890", text: "Hi there"},
		"whatsapp://send/?phone=%2B62+812-3456-7890":                                               {phone: "6281234567890"},
		"whatsapp://send?text=no+number":                                                           {text: "no number", share: true},
		"whatsup://send?text=hi":                                                                   {text: "hi", share: true},
		"whatsapp://chat?code=FrLhsO1aBMEBhOFBtpLwxl":                                              {code: "FrLhsO1aBMEBhOFBtpLwxl"},
		"https://wa.me/message/ABCDEF":                                                             {resolve: "ABCDEF", kind: model.LinkBusiness},
		"https://wa.me/qr/QRCODE12":                                                                {resolve: "QRCODE12", kind: model.LinkContact},
		"https://whatsapp.com/channel/0029VaExample":                                               {resolve: "0029VaExample", kind: model.LinkChannel},
		"https://api.whatsapp.com/send/?text=https%3A%2F%2Fwww.reddit.com%2Fr%2Fx&type=custom_url": {text: "https://www.reddit.com/r/x", share: true},
		// No phone, group, channel, business or QR link.
		"https://wa.me/6281234567890/extra/path":       nil,
		"https://wa.me/notanumber":                     nil,
		"https://wa.me/123":                            nil,
		"https://wa.me.evil.example/6281234567890":     nil,
		"https://wa.me/channel/0029VaExample":          nil,
		"https://wa.me/qr/bad-code!":                   nil,
		"https://wa.me/c/6281234567890":                nil,
		"https://example.com/6281234567890":            nil,
		"https://example.com/send?phone=6281234567890": nil,
		"https://api.whatsapp.com/about":               nil,
		"whatsapp://send?phone=12&text=hello":          nil,
		"whatsapp://send":                              nil,
		"whatsapp://chat?code=bad/../code":             nil,
		"/kick @Gusti":                                 nil,
		"6281234567890":                                nil,
		"":                                             nil,
	} {
		got, ok := parseAppLink(link)
		switch {
		case want == nil && ok:
			t.Errorf("parseAppLink(%q) = %+v, want none", link, got)
		case want != nil && (!ok || got != *want):
			t.Errorf("parseAppLink(%q) = %+v, %v, want %+v", link, got, ok, *want)
		}
	}
}

// TestOpenChatLink checks that a chat link opens the number's chat with
// its text in the composer, without saving the contact.
func TestOpenChatLink(t *testing.T) {
	st := newSlashTest(t, "work")
	st.u.openChatLink("6281255500199", "hello there")
	st.u.applyEvents()
	st.frame()
	if id := st.u.selected.ID; id != "6281255500199@s.whatsapp.net" {
		t.Fatalf("opened %q", id)
	}
	if got := st.text(); got != "hello there" {
		t.Fatalf("composer has %q", got)
	}
}

// TestPendingLinks checks that links left by a second launch open once a
// window shows, and that the file drains.
func TestPendingLinks(t *testing.T) {
	st := newSlashTest(t, "work")
	path := filepath.Join(t.TempDir(), "pending-link")
	os.WriteFile(path, []byte("whatsup://send?phone=6281255500199&text=hai\nnot a link\n"), 0o600)
	h := &host{u: st.u, o: Options{LinkFile: path}}
	h.openLinks()
	st.u.applyEvents()
	st.frame()
	if id := st.u.selected.ID; id != "6281255500199@s.whatsapp.net" {
		t.Fatalf("opened %q", id)
	}
	if got := st.text(); got != "hai" {
		t.Fatalf("composer has %q", got)
	}
	if left, _ := os.ReadFile(path); len(left) != 0 {
		t.Fatalf("pending file kept %q", left)
	}
}
