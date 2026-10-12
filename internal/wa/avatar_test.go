package wa

import (
	"context"
	"testing"

	"github.com/polymorfa/hypermeow/types"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

// TestAvatarRestricted checks that profile pictures download for every
// picture by default, but only for the open chat and its members once
// restricted to open chats.
func TestAvatarRestricted(t *testing.T) {
	ctx := context.Background()
	s, db := testStore(t)
	b := &Backend{ctx: ctx, store: *s}
	mustJID := func(s string) types.JID {
		j, err := types.ParseJID(s)
		if err != nil {
			t.Fatal(err)
		}
		return j
	}
	member := mustJID("111@s.whatsapp.net")
	if err := s.ensureChat(ctx, db, "g@g.us", true, "g"); err != nil {
		t.Fatal(err)
	}
	if err := addMember(ctx, db, "g@g.us", member, types.EmptyJID); err != nil {
		t.Fatal(err)
	}
	if !b.avatarAllowed("g@g.us") || !b.avatarAllowed("111@s.whatsapp.net") {
		t.Fatal("open fetching blocked by default")
	}
	b.SetPref(model.PrefAvatars, model.AvatarsOpen)
	b.openChat = "g@g.us"
	if !b.avatarAllowed("g@g.us") {
		t.Error("the open chat's picture may not download")
	}
	if !b.avatarAllowed("111@s.whatsapp.net") {
		t.Error("an open group member's picture may not download")
	}
	if b.avatarAllowed("999@s.whatsapp.net") {
		t.Error("an outsider's picture downloads in open-chats-only mode")
	}
	b.SetPref(model.PrefAvatars, "")
	if !b.avatarAllowed("999@s.whatsapp.net") {
		t.Error("automatic mode still restricted")
	}
}
