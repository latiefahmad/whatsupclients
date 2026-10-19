package ui

import (
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// Chat links (wa.me, api.whatsapp.com/send, whatsup://) open a chat with
// a phone number without saving the contact; parseAppLink (applink.go)
// reads them.

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
