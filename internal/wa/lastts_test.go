package wa

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

// order returns the chats' jids as the list shows them, newest first.
func order(t *testing.T, s *msgStore) []string {
	t.Helper()
	chats, err := s.chats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, len(chats))
	for i, c := range chats {
		out[i] = c.ID
	}
	return out
}

func assertOrder(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("order: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order: got %v, want %v", got, want)
		}
	}
}

// TestDeleteMovesChatDown checks that deleting a chat's newest message moves
// the chat down to where its remaining newest message belongs. last_ts only
// ever went up, so a chat whose last message was deleted stayed at the top.
func TestDeleteMovesChatDown(t *testing.T) {
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

	// "a" has an old message and, today, a new one.
	if _, err := db.ExecContext(ctx, `INSERT INTO wz_chats (jid, last_ts) VALUES ('a', 3000), ('b', 2000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO wz_messages (chat, id, ts) VALUES
		('a', 'old', 1000), ('a', 'new', 3000), ('b', 'b1', 2000)`); err != nil {
		t.Fatal(err)
	}
	assertOrder(t, order(t, s), "a", "b")

	if _, err := s.deleteMessageEvery(ctx, []string{"a"}, "new"); err != nil {
		t.Fatal(err)
	}
	// "a"'s newest left is 1000, so it belongs below "b" at 2000.
	assertOrder(t, order(t, s), "b", "a")
}

// TestClearMovesChatDown checks the same for a clear from another device,
// which deletes a stretch of messages rather than one.
func TestClearMovesChatDown(t *testing.T) {
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

	if _, err := db.ExecContext(ctx, `INSERT INTO wz_chats (jid, last_ts) VALUES ('a', 3000), ('b', 2000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO wz_messages (chat, id, ts) VALUES
		('a', 'old', 1000), ('a', 'mid', 2500), ('a', 'new', 3000), ('b', 'b1', 2000)`); err != nil {
		t.Fatal(err)
	}

	if err := s.clearChatEvery(ctx, []string{"a"}, 2500); err != nil {
		t.Fatal(err)
	}
	// Cleared up to 2500, so only 3000 is left: "a" stays on top.
	assertOrder(t, order(t, s), "a", "b")
}

// TestDeleteChatKeepsRemainingLastTS checks that a delete that keeps newer
// messages lowers last_ts to them, rather than leaving the chat where the
// deleted ones put it.
func TestDeleteChatKeepsRemainingLastTS(t *testing.T) {
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

	if _, err := db.ExecContext(ctx, `INSERT INTO wz_chats (jid, last_ts) VALUES ('a', 3000), ('b', 2000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO wz_messages (chat, id, ts) VALUES
		('a', 'old', 1200), ('a', 'keep', 1500), ('b', 'b1', 2000)`); err != nil {
		t.Fatal(err)
	}

	if err := s.deleteChatEvery(ctx, []string{"a"}, 1200); err != nil {
		t.Fatal(err)
	}
	// "a" keeps only 'keep' at 1500, which is below "b" at 2000.
	assertOrder(t, order(t, s), "b", "a")
}

// TestDeleteLastMessageKeepsChat checks that a chat emptied by a deletion
// keeps its place: WhatsApp still lists it until the chat itself is removed.
func TestDeleteLastMessageKeepsChat(t *testing.T) {
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

	if _, err := db.ExecContext(ctx, `INSERT INTO wz_chats (jid, last_ts) VALUES ('a', 4000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO wz_messages (chat, id, ts) VALUES ('a', 'only', 4000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.deleteMessageEvery(ctx, []string{"a"}, "only"); err != nil {
		t.Fatal(err)
	}

	var lastTS int64
	if err := db.QueryRowContext(ctx, `SELECT last_ts FROM wz_chats WHERE jid = 'a'`).Scan(&lastTS); err != nil {
		t.Fatal(err)
	}
	if lastTS != 4000 {
		t.Errorf("last_ts after emptying the chat: %d, want 4000 kept", lastTS)
	}
	assertOrder(t, order(t, s), "a")
}
