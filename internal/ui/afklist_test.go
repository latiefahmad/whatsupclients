package ui

import (
	"strings"
	"testing"
)

// afkItemRows returns the AFK list's member rows.
func afkItemRows(u *UI) []settingRow {
	var out []settingRow
	for _, sec := range u.settingsRows() {
		for _, row := range sec.rows {
			if strings.HasPrefix(row.key, "afk:item:") {
				out = append(out, row)
			}
		}
	}
	return out
}

func TestAFKListSettings(t *testing.T) {
	st := newSlashTest(t, "rina")
	u := st.u
	u.ShowPage("afklist")
	st.frame()
	if rows := afkItemRows(u); len(rows) != 0 {
		t.Fatalf("starts with %d rows", len(rows))
	}
	s := u.settings.afk
	if s == nil {
		t.Fatal("no AFK settings state")
	}
	// A typed number takes its contact's name; dups and short
	// numbers are refused.
	s.number.SetText("+62 813-5550-1942")
	u.addAFKNumber()
	if l := u.auto.AllowList(); len(l.Members) != 1 || l.Members[0].Name != "Agus Wibowo" {
		t.Fatalf("listed %+v", l.Members)
	}
	// The + button adds too: its click is read before its layout.
	s.number.SetText("+62 857-5550-3310")
	s.add.Click()
	st.frame()
	if l := u.auto.AllowList(); len(l.Members) != 2 || l.Members[1].Name != "Fitri Handayani" {
		t.Fatalf("button listed %+v", l.Members)
	}
	// A name calls the contact.
	s.number.SetText("Yoga")
	u.addAFKNumber()
	if l := u.auto.AllowList(); len(l.Members) != 3 || l.Members[2].Name != "Yoga Pratama" {
		t.Fatalf("name listed %+v", l.Members)
	}
	s.number.SetText("Nobody Here")
	u.addAFKNumber()
	if !strings.Contains(s.message, "No contact named") {
		t.Errorf("unknown message %q", s.message)
	}
	// The field suggests contacts as you type; tapping one adds it.
	s.number.SetText("")
	u.auto.UnallowMember("6281355501942")
	u.auto.UnallowMember("6285755503310")
	u.auto.UnallowMember("Yoga Pratama")
	u.settings.stale = true
	st.frame()
	s.number.SetText("agu")
	st.frame()
	matched := false
	for _, ct := range u.afkPickMatches() {
		if ct.Name == "Agus Wibowo" {
			matched = true
		}
	}
	if !matched {
		t.Fatal("agu matches nothing")
	}
	u.btn("settings:afk:pick:agus").Click()
	st.frame()
	if l := u.auto.AllowList(); len(l.Members) != 1 || l.Members[0].ID != "agus" {
		t.Fatalf("pick listed %+v", l.Members)
	}
	s.number.SetText("+62 813-5550-1942")
	u.addAFKNumber()
	if !strings.Contains(s.message, "already on the list") {
		t.Errorf("dup message %q", s.message)
	}
	s.number.SetText("123")
	u.addAFKNumber()
	if !strings.Contains(s.message, "country code") {
		t.Errorf("short message %q", s.message)
	}
	// The confine switch flips the mode.
	settingRowByKey(t, u, "afk:only").run()
	if !u.auto.AllowList().Only {
		t.Fatal("switch didn't confine")
	}
	// The reason starts being away, then updates it without restarting.
	s.reason.SetText("lunch")
	settingRowByKey(t, u, "afk:reason:save").run()
	aw := u.auto.Away()
	if aw == nil || aw.Reason != "lunch" {
		t.Fatalf("away %+v", aw)
	}
	since := aw.Since
	s.reason.SetText("back at 2")
	settingRowByKey(t, u, "afk:reason:save").run()
	if aw := u.auto.Away(); aw == nil || aw.Reason != "back at 2" || !aw.Since.Equal(since) {
		t.Fatalf("updated %+v", aw)
	}
	// The hours confine replies, then lift again when cleared.
	s.from.SetText("21:00")
	s.to.SetText("07:00")
	settingRowByKey(t, u, "afk:hours:save").run()
	if h := u.auto.Hours(); !h.On || h.From != 21*60 || h.To != 7*60 {
		t.Fatalf("hours %+v", h)
	}
	s.to.SetText("noon")
	settingRowByKey(t, u, "afk:hours:save").run()
	if !strings.Contains(s.message, "hour like") {
		t.Errorf("hours message %q", s.message)
	}
	s.from.SetText("")
	s.to.SetText("")
	settingRowByKey(t, u, "afk:hours:save").run()
	if u.auto.Hours().On {
		t.Fatal("hours didn't lift")
	}
	// Search filters the rows.
	s.search.SetText("agus")
	st.frame()
	if rows := afkItemRows(u); len(rows) != 1 {
		t.Fatalf("search found %d rows", len(rows))
	}
	s.search.SetText("nobody here")
	st.frame()
	if rows := afkItemRows(u); len(rows) != 0 {
		t.Fatalf("search found %d rows", len(rows))
	}
	s.search.SetText("")
	st.frame()
	// Confined hours show in the note.
	u.auto.SetHours(true, 21*60, 7*60)
	u.settings.stale = true
	st.frame()
	found := false
	for _, sec := range u.settingsRows() {
		if strings.Contains(sec.note, "21:00") {
			found = true
		}
	}
	if !found {
		t.Fatal("hours note missing")
	}
	u.auto.SetHours(false, 0, 0)
	// A reply marks the member; removing asks first.
	u.auto.SetAway("lunch")
	w := u.auto.Away()
	w.Told = append(w.Told, "sometext|6281355501942")
	u.settings.stale = true
	st.frame()
	rows := afkItemRows(u)
	var agus settingRow
	for _, row := range rows {
		if strings.Contains(row.title, "Agus") {
			agus = row
		}
	}
	if agus.title == "" || !strings.Contains(agus.sub, "✓") {
		t.Fatalf("told rows %+v", rows)
	}
	agus.run()
	if len(u.dialog.buttons) != 2 || u.dialog.buttons[0].label != "Remove" {
		t.Fatalf("remove dialog %+v", u.dialog.buttons)
	}
	u.dialog.buttons[0].run()
	for _, m := range u.auto.AllowList().Members {
		if strings.Contains(m.Name, "Agus") {
			t.Fatalf("remove left %+v", u.auto.AllowList().Members)
		}
	}
	u.closeDialog()
	u.settingsBack()
	if u.settings.afk != nil {
		t.Fatal("back didn't drop the state")
	}
}
