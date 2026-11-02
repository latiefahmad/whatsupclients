package wa

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	waLog "github.com/polymorfa/hypermeow/util/log"
)

// TestMergeSenderLIDChats checks that a chat a bot opened under a LID nothing
// maps is folded into the chat its phone number is known by, so the same
// message isn't stored twice and doesn't notify twice.
func TestMergeSenderLIDChats(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &msgStore{db: db}
	if err := s.init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := makeLIDMap(ctx, db); err != nil {
		t.Fatal(err)
	}

	const pn = "628567867504"
	const knownLID = "61289787343003"
	const strayLID = "6790846652571"

	for _, q := range []string{
		`INSERT INTO whatsmeow_lid_map (lid, pn) VALUES ('` + knownLID + `', '` + pn + `')`,
		`INSERT INTO wz_chats (jid) VALUES ('` + knownLID + `@lid')`,
		`INSERT INTO wz_chats (jid) VALUES ('` + strayLID + `@lid')`,
		// The conversation lives under the known LID...
		`INSERT INTO wz_messages (chat, id, ts, sender_jid) VALUES
			('` + knownLID + `@lid', 'm1', 100, '` + pn + `@s.whatsapp.net')`,
		// ...and the same message was filed under the stray LID too.
		`INSERT INTO wz_messages (chat, id, ts, sender_jid) VALUES
			('` + strayLID + `@lid', 'm1', 100, '` + pn + `@s.whatsapp.net'),
			('` + strayLID + `@lid', 'm2', 200, '` + pn + `@s.whatsapp.net')`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}

	if err := s.mergeSenderLIDChats(ctx, testLogger{}); err != nil {
		t.Fatal(err)
	}

	if s.hasMessages(ctx, strayLID+"@lid") {
		t.Error("stray chat still holds messages")
	}
	if !s.hasMessages(ctx, knownLID+"@lid") {
		t.Error("known chat lost its messages")
	}
	var n int
	db.QueryRowContext(ctx, `SELECT count(*) FROM wz_messages WHERE chat = ? AND id = 'm1'`,
		knownLID+"@lid").Scan(&n)
	if n != 1 {
		t.Errorf("m1 kept %d times in the known chat, want 1", n)
	}
	db.QueryRowContext(ctx, `SELECT count(*) FROM wz_messages WHERE id = 'm2'`).Scan(&n)
	if n != 1 {
		t.Errorf("m2 kept %d times, want 1", n)
	}
}

// TestMergeSenderLIDChatsSkipsMapped checks that a chat under a LID that does
// have a mapping is left alone: it is an identity of its own, not a stray.
func TestMergeSenderLIDChatsSkipsMapped(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &msgStore{db: db}
	if err := s.init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := makeLIDMap(ctx, db); err != nil {
		t.Fatal(err)
	}

	const pn = "628111"
	const mapped = "111@lid"
	const other = "222@lid"
	for _, q := range []string{
		`INSERT INTO whatsmeow_lid_map (lid, pn) VALUES ('111', '` + pn + `')`,
		`INSERT INTO whatsmeow_lid_map (lid, pn) VALUES ('222', '628999')`,
		`INSERT INTO wz_chats (jid) VALUES ('` + mapped + `'), ('` + other + `')`,
		`INSERT INTO wz_messages (chat, id, ts, sender_jid) VALUES
			('` + mapped + `', 'a', 1, '` + pn + `@s.whatsapp.net'),
			('` + other + `', 'b', 1, '` + pn + `@s.whatsapp.net')`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}

	if err := s.mergeSenderLIDChats(ctx, testLogger{}); err != nil {
		t.Fatal(err)
	}
	if !s.hasMessages(ctx, other) {
		t.Error("a chat under a mapped LID was folded away")
	}
}

// TestMergeSenderLIDChatsOnce checks the merge runs once.
func TestMergeSenderLIDChatsOnce(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &msgStore{db: db}
	if err := s.init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := makeLIDMap(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := s.mergeSenderLIDChats(ctx, testLogger{}); err != nil {
		t.Fatal(err)
	}
	if s.meta(ctx, "sender_lid_chats_merged") == "" {
		t.Error("marker not set")
	}
}

type testLogger struct{}

func (testLogger) Warnf(string, ...any)    {}
func (testLogger) Errorf(string, ...any)   {}
func (testLogger) Infof(string, ...any)    {}
func (testLogger) Debugf(string, ...any)   {}
func (testLogger) Sub(string) waLog.Logger { return testLogger{} }

// makeLIDMap creates hypermeow's own table, which its device store makes and
// this package only reads.
func makeLIDMap(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS whatsmeow_lid_map (
		lid TEXT PRIMARY KEY, pn TEXT UNIQUE NOT NULL)`)
	return err
}
