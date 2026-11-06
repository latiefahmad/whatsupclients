package wa

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/polymorfa/hypermeow"
	"github.com/polymorfa/hypermeow/store/sqlstore"
	"github.com/polymorfa/hypermeow/types"
	waLog "github.com/polymorfa/hypermeow/util/log"
)

// testDeviceJID keys the test session stores. Nothing is written to the
// device table, so it only has to be a valid JID.
var testDeviceJID = types.NewJID("1234", types.DefaultUserServer)

// clientWithStore makes a client whose app state store is real, the way the
// app's own device store is: NewClient leaves Store nil, and both this
// package and hypermeow dereference it.
func clientWithStore(t *testing.T, db *sql.DB) *whatsmeow.Client {
	t.Helper()
	if _, err := db.Exec(`PRAGMA foreign_keys(1)`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	container := sqlstore.NewWithDB(db, "sqlite", waLog.Noop)
	if err := container.Upgrade(context.Background()); err != nil {
		t.Fatalf("upgrade session database: %v", err)
	}
	device := container.NewDevice()
	// NewDevice makes the keys but leaves every store nil; the app's own
	// device is loaded from the database, which sets them.
	device.SetAllStores(sqlstore.NewSQLStore(container, testDeviceJID))
	return whatsmeow.NewClient(device, waLog.Noop)
}

// TestPatchNotSentWithNoState checks that a change isn't sent into a
// collection this device holds no state for. Every such send is answered
// 409, and the conflicting patches can't be applied either ("missing value
// MAC of previous SET operation"), so the change was lost and had no chance
// of arriving. It stays queued instead.
func TestPatchNotSentWithNoState(t *testing.T) {
	s, db := testStore(t)
	log := &captureLogger{}
	b := &Backend{store: *s, ctx: context.Background(), log: log, db: db}
	b.cli = clientWithStore(t, db)

	// Nothing was ever stored for regular_high: version 0, empty hash.
	p := deletePatch("c@s.whatsapp.net", "m1")
	b.sendAppStateNow(b.cli, p)

	// The guard's own line is what tells this path from the ordinary
	// failure one: without the guard the change is still queued, so the
	// queue alone can't tell them apart.
	guarded := false
	for _, line := range log.lines {
		if strings.Contains(line, "has no state; kept") {
			guarded = true
		}
	}
	if !guarded {
		t.Errorf("sent into an empty collection; log: %q", log.lines)
	}
	if n := b.pendingLen(); n != 1 {
		t.Fatalf("pending: %d, want 1", n)
	}
}
