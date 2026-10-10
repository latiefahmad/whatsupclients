package ui

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/latiefahmad/whatsupclients/internal/auto"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// afkListSettings is Settings > AFK list: the away message, who the
// AFK reply goes to, and when. The list itself lives in the auto
// backend's prefs (auto.AFKAllow), so this only holds the editors.
type afkListSettings struct {
	search, number, reason, from, to widget.Editor
	query, message                   string
	filled                           bool
	add                              widget.Clickable
}

func newAFKListSettings() *afkListSettings {
	s := &afkListSettings{}
	s.search.SingleLine = true
	s.number.SingleLine = true
	s.reason.SingleLine = true
	s.from.SingleLine = true
	s.to.SingleLine = true
	return s
}

// afkListSettingsRows draws the AFK list page: the away message, the
// confine switch, a search box, a name or number to add, a row per
// listed contact, and the reply hours.
func (u *UI) afkListSettingsRows() []settingsSection {
	s := u.settings.afk
	if s == nil {
		s = newAFKListSettings()
		u.settings.afk = s
	}
	if !s.filled {
		s.filled = true
		if u.auto != nil {
			if w := u.auto.Away(); w != nil {
				s.reason.SetText(w.Reason)
			}
			if h := u.auto.Hours(); h.On {
				s.from.SetText(auto.FormatHour(h.From))
				s.to.SetText(auto.FormatHour(h.To))
			}
		}
	}
	away := u.auto != nil && u.auto.Away() != nil
	msg := settingsSection{title: "Away message", rows: []settingRow{
		{key: "afk:reason", kind: setCustom, w: func(gtx C) D {
			for {
				ev, ok := s.reason.Update(gtx)
				if !ok {
					break
				}
				if _, ok := ev.(widget.SubmitEvent); ok {
					u.saveAFKReason()
				}
			}
			return layout.Inset{Left: 24, Right: 24, Bottom: 10}.Layout(gtx, func(gtx C) D {
				return u.searchField(gtx, &s.reason, "Why you're away")
			})
		}},
	}}
	if away {
		msg.rows = append(msg.rows, settingRow{key: "afk:reason:save", ic: icTick, title: "Update message",
			run: u.saveAFKReason})
	} else {
		msg.rows = append(msg.rows, settingRow{key: "afk:reason:save", ic: icTick, title: "Turn on AFK",
			run: u.saveAFKReason})
	}
	sec := settingsSection{title: "AFK list"}
	a := u.auto
	only := a != nil && a.AllowList().Only
	sec.rows = append(sec.rows, settingRow{key: "afk:only", kind: setToggle,
		title: "Reply only to listed contacts",
		sub:   "While you're away, everyone else gets no reply",
		on:    only, run: func() {
			if u.auto != nil {
				u.auto.SetAllowOnly(!only)
			}
		}})
	sec.rows = append(sec.rows, settingRow{key: "afk:search", kind: setCustom, w: func(gtx C) D {
		return layout.Inset{Left: 24, Right: 24, Bottom: 10}.Layout(gtx, func(gtx C) D {
			return u.searchField(gtx, &s.search, "Search the list")
		})
	}})
	sec.rows = append(sec.rows, settingRow{key: "afk:add", kind: setCustom, w: func(gtx C) D {
		for {
			ev, ok := s.number.Update(gtx)
			if !ok {
				break
			}
			if _, ok := ev.(widget.SubmitEvent); ok {
				u.addAFKNumber()
			}
		}
		// Before the layouts below: Layout drains clicks.
		if s.add.Clicked(gtx) {
			u.addAFKNumber()
		}
		matches := u.afkPickMatches()
		for _, ct := range matches {
			ct := ct
			if u.btn("settings:afk:pick:" + ct.ID).Clicked(gtx) {
				if u.auto != nil && u.auto.AllowMember(ct.ID, ct.Name) {
					s.message = "Added " + ct.Name + "."
					s.number.SetText("")
				} else {
					s.message = ct.Name + " is already on the list."
				}
				u.settings.stale = true
			}
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: 24, Right: 24, Bottom: 10}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx C) D {
							return u.searchField(gtx, &s.number, "Name or phone number")
						}),
						layout.Rigid(layout.Spacer{Width: 12}.Layout),
						layout.Rigid(func(gtx C) D {
							return u.iconButton(gtx, &s.add, icAdd, 40, 24, u.pal.IconStrong)
						}),
					)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: 24, Right: 24}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, u.afkPickRows(matches)...)
				})
			}),
		)
	}})
	members := []auto.AFKMember{}
	if a != nil {
		members = a.AllowList().Members
	}
	told := u.afkToldMembers(members)
	q := strings.ToLower(trimSpace(s.query))
	for _, m := range members {
		if q != "" && !strings.Contains(strings.ToLower(m.Name+" "+m.ID), q) {
			continue
		}
		sub := u.afkMemberSub(m)
		if told[m.ID] {
			sub += " · ✓ replied"
		}
		m := m
		sec.rows = append(sec.rows, settingRow{key: "afk:item:" + m.ID, kind: setContact,
			id: m.ID, title: m.Name, sub: sub, run: func() {
				u.confirm("Remove "+m.Name+" from the AFK list?", "",
					dialogButton{label: "Remove", primary: true, run: func() {
						if u.auto != nil {
							u.auto.UnallowMember(m.ID)
						}
						u.settings.stale = true
					}})
			}})
	}
	hrs := settingsSection{title: "Reply hours", rows: []settingRow{
		{key: "afk:hours", kind: setCustom, w: func(gtx C) D {
			for _, ed := range []*widget.Editor{&s.from, &s.to} {
				for {
					ev, ok := ed.Update(gtx)
					if !ok {
						break
					}
					if _, ok := ev.(widget.SubmitEvent); ok {
						u.saveAFKHours()
					}
				}
			}
			return layout.Inset{Left: 24, Right: 24, Bottom: 10}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx C) D {
						return u.searchField(gtx, &s.from, "From 21:00")
					}),
					layout.Rigid(layout.Spacer{Width: 12}.Layout),
					layout.Flexed(1, func(gtx C) D {
						return u.searchField(gtx, &s.to, "To 07:00")
					}),
				)
			})
		}},
		{key: "afk:hours:save", ic: icTick, title: "Save hours", sub: "Empty hours reply all day",
			run: u.saveAFKHours},
	}}
	note := "While you're away, the reply goes to everyone, or only to these contacts when confined above."
	switch {
	case s.message != "":
		note = s.message
	case len(members) == 0:
		note += " No listed contacts yet: add a name or number above."
	}
	if u.auto != nil {
		if h := u.auto.Hours(); h.On {
			note += " Replies go " + auto.FormatHour(h.From) + "–" + auto.FormatHour(h.To) + "."
		}
	}
	return []settingsSection{msg, sec, hrs, {note: note}}
}

// saveAFKReason starts being away with the message, or updates it.
func (u *UI) saveAFKReason() {
	s := u.settings.afk
	if s == nil || u.auto == nil {
		return
	}
	reason := trimSpace(s.reason.Text())
	if u.auto.Away() == nil {
		u.auto.SetAway(reason)
		s.message = "You're away now."
	} else {
		u.auto.SetAwayReason(reason)
		s.message = "Message updated."
	}
	u.settings.stale = true
}

// saveAFKHours confines replies to the hours, or lifts them when both
// are empty.
func (u *UI) saveAFKHours() {
	s := u.settings.afk
	if s == nil || u.auto == nil {
		return
	}
	fromT, toT := trimSpace(s.from.Text()), trimSpace(s.to.Text())
	if fromT == "" && toT == "" {
		u.auto.SetHours(false, 0, 0)
		s.message = "Replies go at all hours."
		u.settings.stale = true
		return
	}
	if fromT == "" || toT == "" {
		s.message = "Fill both hours, or neither for all day."
		u.settings.stale = true
		return
	}
	from, err := auto.ParseHour(fromT)
	if err != nil {
		s.message = "First hour: " + strings.ToLower(err.Error()) + "."
		u.settings.stale = true
		return
	}
	to, err := auto.ParseHour(toT)
	if err != nil {
		s.message = "Second hour: " + strings.ToLower(err.Error()) + "."
		u.settings.stale = true
		return
	}
	u.auto.SetHours(true, from, to)
	s.message = "Replies go " + auto.FormatHour(from) + "–" + auto.FormatHour(to) + "."
	if from > to {
		s.message += " Overnight, past midnight."
	}
	u.settings.stale = true
}

// afkPickMatches are the saved contacts the number field's text finds
// (by name or number), minus the listed ones, at most a few.
func (u *UI) afkPickMatches() []*model.Contact {
	s := u.settings.afk
	if s == nil {
		return nil
	}
	q := trimSpace(s.number.Text())
	if q == "" {
		return nil
	}
	listed := map[string]bool{}
	if u.auto != nil {
		for _, m := range u.auto.AllowList().Members {
			listed[m.ID] = true
			if d := auto.Digits(m.ID); len(d) >= 7 {
				listed[d] = true
			}
		}
	}
	ql, qd := strings.ToLower(q), auto.Digits(q)
	var out []*model.Contact
	for _, ct := range u.backend.Contacts() {
		if listed[ct.ID] {
			continue
		}
		hit := strings.Contains(strings.ToLower(ct.Name), ql)
		if !hit && len(qd) >= 3 && strings.Contains(auto.Digits(ct.Phone), qd) {
			hit = true
		}
		if !hit {
			continue
		}
		out = append(out, ct)
		if len(out) == 6 {
			break
		}
	}
	return out
}

// afkPickRows draws the contacts to tap into the list.
func (u *UI) afkPickRows(matches []*model.Contact) []layout.FlexChild {
	var rows []layout.FlexChild
	for _, ct := range matches {
		ct := ct
		rows = append(rows, layout.Rigid(func(gtx C) D {
			return clickable(gtx, u.btn("settings:afk:pick:"+ct.ID), func(gtx C) D {
				return layout.Inset{Top: 7, Bottom: 7}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return u.avatar(gtx, ct.ID, ct.Name, false, 36)
						}),
						layout.Rigid(layout.Spacer{Width: 12}.Layout),
						layout.Flexed(1, func(gtx C) D {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(u.label(15, ct.Name, u.pal.Text, labelOpts{maxLines: 1}).Layout),
								layout.Rigid(func(gtx C) D {
									if ct.Phone == "" {
										return D{}
									}
									return u.label(12.5, ct.Phone, u.pal.TextSecondary, labelOpts{maxLines: 1}).Layout(gtx)
								}),
							)
						}),
					)
				})
			})
		}))
	}
	return rows
}
func (u *UI) addAFKNumber() {
	s := u.settings.afk
	if s == nil || u.auto == nil {
		return
	}
	text := trimSpace(s.number.Text())
	id, name, err := auto.ResolvePerson(u.backend, nil, text)
	if err != nil {
		s.message = strings.ToUpper(err.Error()[:1]) + err.Error()[1:] + "."
		u.settings.stale = true
		return
	}
	if !u.auto.AllowMember(id, name) {
		s.message = name + " is already on the list."
	} else {
		s.message = "Added " + name + "."
		s.number.SetText("")
	}
	u.settings.stale = true
}

// afkMemberSub is a listed contact's number, when known.
func (u *UI) afkMemberSub(m auto.AFKMember) string {
	for _, ct := range u.backend.Contacts() {
		if ct.ID == m.ID || (len(auto.Digits(ct.Phone)) >= 7 && auto.Digits(ct.Phone) == auto.Digits(m.ID)) {
			if ct.Phone != "" {
				return ct.Phone
			}
			break
		}
	}
	if m.ID != m.Name {
		return m.ID
	}
	return ""
}

// afkToldMembers marks the members who got the AFK reply so far, in
// any ID form the reply recorded.
func (u *UI) afkToldMembers(members []auto.AFKMember) map[string]bool {
	out := map[string]bool{}
	if u.auto == nil {
		return out
	}
	w := u.auto.Away()
	if w == nil {
		return out
	}
	set := map[string]bool{}
	for _, k := range w.Told {
		for _, part := range strings.Split(k, "|") {
			set[part] = true
			if d := auto.Digits(part); len(d) >= 7 {
				set[d] = true
			}
			if ck := u.auto.AllowKey(part); ck != "" {
				set[ck] = true
			}
		}
	}
	for _, m := range members {
		if set[m.ID] {
			out[m.ID] = true
			continue
		}
		if d := auto.Digits(m.ID); len(d) >= 7 && set[d] {
			out[m.ID] = true
			continue
		}
		if ck := u.auto.AllowKey(m.ID); ck != "" && set[ck] {
			out[m.ID] = true
		}
	}
	return out
}
