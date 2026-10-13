package ui

import (
	"net/url"
	"strings"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

// ParseChatLink reads a phone number (digits with the country code) and a
// message text from a chat link, for opening the chat without saving the
// contact:
//
//	whatsup://send?phone=62812...&text=hello
//	whatsup://62812...?text=hello
//	https://wa.me/62812...?text=hello
//	https://api.whatsapp.com/send?phone=62812...&text=hello
//
// It reports false when s is no chat link.
func ParseChatLink(s string) (phone, text string, ok bool) {
	s = strings.TrimSpace(strings.Trim(s, `"'`))
	u, err := url.Parse(s)
	if err != nil {
		return "", "", false
	}
	q := u.Query()
	switch strings.ToLower(u.Scheme) {
	case "whatsup":
		if u.Host == "send" {
			if phone = phoneDigits(q.Get("phone")); phone != "" {
				return phone, q.Get("text"), true
			}
			return "", "", false
		}
		if u.Opaque != "" {
			// whatsup:<digits>?text=...
			if phone = phoneDigits(u.Opaque); phone != "" {
				return phone, q.Get("text"), true
			}
			return "", "", false
		}
		// whatsup://<digits>?text=... (the host holds the number).
		if phone = phoneDigits(u.Host); phone != "" {
			return phone, q.Get("text"), true
		}
		return "", "", false
	case "http", "https":
		host := strings.ToLower(u.Host)
		switch {
		case host == "wa.me" || strings.HasSuffix(host, ".wa.me"):
			if phone = phoneDigits(strings.Trim(u.Path, "/")); phone != "" {
				return phone, q.Get("text"), true
			}
			return "", "", false
		case host == "api.whatsapp.com" && strings.ToLower(strings.Trim(u.Path, "/")) == "send":
			if phone = phoneDigits(q.Get("phone")); phone != "" {
				return phone, q.Get("text"), true
			}
			return "", "", false
		}
	}
	return "", "", false
}

// openChatLink opens the chat with a phone number (digits with the
// country code) once the backend has looked it up (linkPhoneEvent),
// prefilling the composer's text.
func (u *UI) openChatLink(phone, text string) {
	u.linkPhone, u.linkText = phone, text
	u.backend.LookupPhone(phone)
}

// linkPhoneEvent answers openChatLink, and reports whether e was for it.
func (u *UI) linkPhoneEvent(e model.PhoneEvent) bool {
	if u.linkPhone == "" || e.Phone != u.linkPhone {
		return false
	}
	text := u.linkText
	u.linkPhone, u.linkText = "", ""
	switch {
	case e.Err != "":
		u.toast(e.Err)
	case e.ID == "":
		u.toast("+" + e.Phone + " isn't on WhatsApp.")
	default:
		name := first(e.Name, "+"+e.Phone)
		if !u.openChatByID(e.ID) {
			u.openDirect(e.ID, name)
		}
		if text != "" {
			ed := &u.conv.composer
			ed.SetText(text)
			ed.SetCaret(ed.Len(), ed.Len())
			u.requestFocus(ed)
		}
	}
	return true
}
