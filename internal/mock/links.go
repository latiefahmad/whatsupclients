package mock

import "github.com/latiefahmad/whatsupclients/internal/model"

// Demo codes of the links only the server resolves: a suggested channel
// (bola), a followed one (techdaily), a business's message link (Vivian)
// and a contact's QR link (Arman).
const (
	LinkChannelSuggested = "DemoChannelBola"
	LinkChannelFollowed  = "DemoChannelTech"
	LinkBusiness         = "DemoBusinessVivy"
	LinkContact          = "DemoContactArman"
)

// ResolveLink answers for the demo codes.
func (b *Backend) ResolveLink(kind model.LinkKind, code string) {
	ev := model.LinkEvent{Kind: kind, Code: code}
	channel := func(id, desc string) {
		for _, list := range [][]*model.Channel{b.channels, b.suggested} {
			for _, ch := range list {
				if ch.ID == id {
					c := *ch
					ev.ID, ev.Name, ev.Channel, ev.Description = id, ch.Name, &c, desc
				}
			}
		}
	}
	switch {
	case kind == model.LinkChannel && code == LinkChannelSuggested:
		channel("bola@newsletter", "Scores, transfers and match highlights from the national league, every matchday.")
	case kind == model.LinkChannel && code == LinkChannelFollowed:
		channel("techdaily@newsletter", "")
	case kind == model.LinkBusiness && code == LinkBusiness:
		ev.ID, ev.Name, ev.Text = "vivy@lid", "Vivian", "Hi! I'd like to know more about your plans."
	case kind == model.LinkContact && code == LinkContact:
		ev.ID, ev.Name = "arman", "Arman Maulana"
	}
	if ev.ID == "" {
		ev.Err = "This link doesn't exist or was reset."
	}
	b.emit(ev)
}
