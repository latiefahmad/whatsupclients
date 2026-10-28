package wa

import (
	"context"
	"database/sql"
	"testing"
	"time"

	waLog "github.com/polymorfa/hypermeow/util/log"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

func TestMergeChat(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	b := &Backend{ctx: context.Background(), db: db, store: msgStore{db: db}, log: waLog.Noop, dataDir: t.TempDir()}
	if err := b.store.init(b.ctx); err != nil {
		t.Fatal(err)
	}
	ctx := b.ctx
	pn, lid := "6281234567890@s.whatsapp.net", "123456789012345@lid"
	now := time.Now()

	// The old phone-number row: messages and state.
	if err := b.store.ensureChat(ctx, db, pn, false, "Old"); err != nil {
		t.Fatal(err)
	}
	_ = b.store.setField(ctx, pn, "unread", 2)
	_ = b.store.setField(ctx, pn, "archived", true)
	for i, id := range []string{"m1", "m2"} {
		if err := b.store.putMessage(ctx, db, storedMsg{Message: &model.Message{ID: id, ChatID: pn,
			FromMe: false, Time: now.Add(-time.Duration(2-i) * time.Minute), Media: model.MediaNone, Text: "hi"}}); err != nil {
			t.Fatal(err)
		}
	}

	// The LID row: later, one newer message, named differently.
	if err := b.store.ensureChat(ctx, db, lid, false, "New"); err != nil {
		t.Fatal(err)
	}
	if err := b.store.putMessage(ctx, db, storedMsg{Message: &model.Message{ID: "m3", ChatID: lid,
		FromMe: false, Time: now, Media: model.MediaNone, Text: "later"}}); err != nil {
		t.Fatal(err)
	}

	if err := b.store.mergeChat(ctx, db, pn, lid); err != nil {
		t.Fatal(err)
	}

	if _, ok := b.store.chat(ctx, pn); ok {
		t.Error("the phone-number row stayed")
	}
	c, ok := b.store.chat(ctx, lid)
	if !ok {
		t.Fatal("the LID row went away")
	}
	if c.Unread != 2 {
		t.Errorf("unread = %d, want 2 (both from the phone-number row)", c.Unread)
	}
	if !c.Archived {
		t.Error("archived didn't merge")
	}
	if c.Name != "New" {
		t.Errorf("name = %q, want the target's kept (New)", c.Name)
	}
	msgs, err := b.store.messages(ctx, lid, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 3 {
		t.Fatalf("messages = %d, want 3", len(msgs))
	}
	// The old row's name fills in only when the target has none; check a
	// merge into a nameless chat.
	lid2 := "999988887777666@lid"
	if err := b.store.putMessage(ctx, db, storedMsg{Message: &model.Message{ID: "m4", ChatID: lid2,
		FromMe: false, Time: now, Media: model.MediaNone}}); err != nil {
		t.Fatal(err)
	}
	if err := b.store.mergeChat(ctx, db, "never-existed@s.whatsapp.net", lid2); err != nil {
		t.Fatal(err)
	}
}
