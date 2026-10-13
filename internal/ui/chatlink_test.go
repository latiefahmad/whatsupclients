package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseChatLink(t *testing.T) {
	for _, c := range []struct {
		in, phone, text string
		ok              bool
	}{
		{"whatsup://send?phone=6281234567890", "6281234567890", "", true},
		{"whatsup://send?phone=6281234567890&text=hi%20there", "6281234567890", "hi there", true},
		{"whatsup://6281234567890?text=hello", "6281234567890", "hello", true},
		{"whatsup:6281234567890", "6281234567890", "", true},
		{"https://wa.me/6281234567890", "6281234567890", "", true},
		{"https://wa.me/6281234567890?text=order%20status%3F", "6281234567890", "order status?", true},
		{"http://wa.me/6281234567890", "6281234567890", "", true},
		{"https://api.whatsapp.com/send?phone=6281234567890&text=hi", "6281234567890", "hi", true},
		{`"https://wa.me/6281234567890"`, "6281234567890", "", true},
		{"https://wa.me/6281234567890/extra/path", "", "", false},
		{"https://wa.me/notanumber", "", "", false},
		{"https://wa.me/123", "", "", false},
		{"https://example.com/6281234567890", "", "", false},
		{"https://api.whatsapp.com/about", "", "", false},
		{"whatsup://send?text=hi", "", "", false},
		{"/kick @Gusti", "", "", false},
		{"6281234567890", "", "", false},
		{"", "", "", false},
	} {
		phone, text, ok := ParseChatLink(c.in)
		if phone != c.phone || text != c.text || ok != c.ok {
			t.Errorf("ParseChatLink(%q) = %q, %q, %v; want %q, %q, %v",
				c.in, phone, text, ok, c.phone, c.text, c.ok)
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
