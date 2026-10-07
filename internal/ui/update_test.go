package ui

import (
	"strings"
	"testing"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/update"
)

func TestUpdateRows(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	u.applyEvents()
	u.ShowPage("help")
	if r := settingRowByKey(t, u, "version"); r.sub != "Development build" {
		t.Errorf("a build without a version shows %q", r.sub)
	}
	for _, sec := range u.settingsRows() {
		for _, r := range sec.rows {
			if r.key == "update" {
				t.Error("a development build offers to update")
			}
		}
	}

	h := &host{o: Options{Version: "v0.9.0", Relaunch: []string{}}, reqs: make(chan request, 16)}
	h.upd.h, h.upd.exe = h, "WhatsUpClients.exe"
	u.host = h
	rows := func() map[string]settingRow {
		u.settings.stale = true
		m := map[string]settingRow{}
		for _, sec := range u.settingsRows() {
			for _, r := range sec.rows {
				m[r.key] = r
			}
		}
		return m
	}
	m := rows()
	if m["version"].sub != "v0.9.0" || m["update"].title != "Check for updates" || m["update"].run == nil {
		t.Fatalf("release build rows: version %q, update %q", m["version"].sub, m["update"].title)
	}
	h.upd.step = updLatest
	if r := rows()["update"]; !strings.Contains(r.title, "up to date") || r.run == nil {
		t.Errorf("up to date row: %q", r.title)
	}
	if _, ok := rows()["whatsnew"]; ok {
		t.Error("What's new shows without a newer release")
	}
}

func TestUpdateOffer(t *testing.T) {
	u := New(mock.New())
	u.Start(func() {})
	h := &host{o: Options{Version: "v0.9.0", Relaunch: []string{}}, reqs: make(chan request, 16)}
	h.b = u.backend
	h.upd.h, h.upd.exe = h, "WhatsUpClients.exe"
	u.host = h

	// Nothing offered while no release is known.
	u.offerUpdate()
	if u.dialog.isOpen() {
		t.Fatal("offered without a release")
	}

	h.upd.step, h.upd.rel = updAvailable, &update.Release{Version: "v0.10.0"}
	u.offerUpdate()
	if !u.dialog.isOpen() || !strings.Contains(u.dialog.title, "v0.10.0") || len(u.dialog.buttons) != 2 {
		t.Fatalf("offer: %+v", u.dialog)
	}

	// Closing it without answering doesn't open it again.
	u.dialog = dialogState{}
	u.offerUpdate()
	if u.dialog.isOpen() {
		t.Fatal("offered again after dismissing it")
	}

	// "Later" skips the version, even across a fresh check.
	h.upd.offered = ""
	h.upd.step, h.upd.rel = updAvailable, &update.Release{Version: "v0.10.0"}
	u.dialog = dialogState{kind: dialogConfirm, title: "Update to v0.10.0 is available",
		buttons: []dialogButton{{label: "Later", run: func() { h.b.SetPref(prefUpdateSkipped, "v0.10.0") }}}}
	u.dialog.buttons[0].run()
	u.dialog = dialogState{}
	u.offerUpdate()
	if u.dialog.isOpen() {
		t.Fatal("offered a skipped version again")
	}

	// A newer release is offered despite the skip.
	h.upd.offered = ""
	h.upd.step, h.upd.rel = updAvailable, &update.Release{Version: "v0.11.0"}
	u.offerUpdate()
	if !u.dialog.isOpen() || !strings.Contains(u.dialog.title, "v0.11.0") {
		t.Fatalf("newer release not offered: %+v", u.dialog)
	}
}
