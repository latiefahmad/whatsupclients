package wa

import (
	"context"
	"slices"
	"testing"

	"github.com/polymorfa/hypermeow/types"
)

// TestChatKeysKeepsTheEventJID checks that a deletion under either identity
// names both keys. A phone number only maps to a LID once hypermeow knows
// it, so a handler that keys off the LID alone misses messages stored under
// the number (or the other way round) and leaves them on this device.
func TestChatKeysKeepsTheEventJID(t *testing.T) {
	ctx := context.Background()
	b := &Backend{} // no client: no mapping to look up

	for _, tc := range []struct {
		name string
		jid  string
		want string
	}{
		{"a LID", "7907101929672@lid", "7907101929672@lid"},
		{"a phone number", "6285733408212@s.whatsapp.net", "6285733408212@s.whatsapp.net"},
		{"a device JID", "6285733408212:16@s.whatsapp.net", "6285733408212@s.whatsapp.net"},
		{"a group", "120363420223782588@g.us", "120363420223782588@g.us"},
	} {
		j, err := types.ParseJID(tc.jid)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		keys := b.chatKeys(ctx, j)
		if !slices.Contains(keys, tc.want) {
			t.Errorf("%s: keys %v don't hold %q", tc.name, keys, tc.want)
		}
		if len(keys) != 1 {
			t.Errorf("%s: %d keys without a mapping, want 1: %v", tc.name, len(keys), keys)
		}
	}
}

// TestChatKeysDedupes checks that a key isn't listed twice: canonical()
// already returns the LID for a phone number hypermeow has mapped.
func TestChatKeysDedupes(t *testing.T) {
	ctx := context.Background()
	b := &Backend{}
	j, _ := types.ParseJID("6285733408212:16@s.whatsapp.net")
	keys := b.chatKeys(ctx, j)
	seen := map[string]bool{}
	for _, k := range keys {
		if seen[k] {
			t.Errorf("%q listed twice: %v", k, keys)
		}
		seen[k] = true
	}
}

// TestChatKeysDelete checks that the keys chatKeys returns are the ones
// deleteMessageEvery deletes by: a copy under the LID and one under the
// phone number both go, however the event named the chat.
func TestChatKeysDelete(t *testing.T) {
	ctx := context.Background()
	s, _ := testStore(t)

	const lid, pn = "7907101929672@lid", "6285733408212@s.whatsapp.net"
	for _, chat := range []string{lid, pn} {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO wz_chats (jid) VALUES (?)`, chat); err != nil {
			t.Fatal(err)
		}
		if _, err := s.db.ExecContext(ctx, `INSERT INTO wz_messages (chat, id, ts) VALUES (?, 'm1', 1)`, chat); err != nil {
			t.Fatal(err)
		}
	}
	var before int
	s.db.QueryRowContext(ctx, `SELECT count(*) FROM wz_messages WHERE id = 'm1'`).Scan(&before)
	if before != 2 {
		t.Fatalf("set up %d copies, want 2", before)
	}
	// A mapping the handler would have looked up, as hypermeow holds it.
	keys := []string{lid, pn}
	n, err := s.deleteMessageEvery(ctx, keys, "m1")
	if err != nil {
		t.Fatal(err)
	}
	// The count is how a caller tells "deleted it" from "was never stored
	// here", so it must count the message, not the last table touched.
	if n != 2 {
		t.Errorf("deleted %d copies, want 2: the count must follow wz_messages", n)
	}
	var left int
	s.db.QueryRowContext(ctx, `SELECT count(*) FROM wz_messages WHERE id = 'm1'`).Scan(&left)
	if left != 0 {
		t.Errorf("%d copies of the message left", left)
	}
}

func mustJID(t *testing.T, s string) types.JID {
	t.Helper()
	j, err := types.ParseJID(s)
	if err != nil {
		t.Fatal(err)
	}
	return j
}
