package wa

import (
	"bytes"
	"testing"
	"time"

	"github.com/polymorfa/hypermeow/proto/waCommon"
	"github.com/polymorfa/hypermeow/proto/waE2E"
	"github.com/polymorfa/hypermeow/types"
	"github.com/polymorfa/hypermeow/types/events"
	"google.golang.org/protobuf/proto"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

const (
	reactChat   = "123@g.us"
	reactParent = "PARENT"
	reactChild  = "CHILD"
)

// reactArrive has message id come as it would from the network, at unix
// second ts.
func reactArrive(t *testing.T, b *Backend, id string, ts int64, m *waE2E.Message) {
	t.Helper()
	jid := types.NewJID("123", types.GroupServer)
	e := &events.Message{Info: types.MessageInfo{MessageSource: types.MessageSource{Chat: jid, Sender: types.NewJID("9", types.HiddenUserServer), IsGroup: true},
		ID: id, Timestamp: time.Unix(ts, 0)}, Message: m, RawMessage: m}
	if err := b.store.ensureChat(b.ctx, b.db, reactChat, true, "g"); err != nil {
		t.Fatal(err)
	}
	p, ok := b.parse(b.ctx, e)
	if !ok {
		t.Fatalf("%s wasn't parsed", id)
	}
	b.apply(&e.Info, p)
}

// reactNewMessages counts the message events of new messages so far, and
// empties the queue.
func reactNewMessages(b *Backend) int {
	n := 0
	for _, e := range b.Poll() {
		if m, ok := e.(model.MessageEvent); ok && m.New {
			n++
		}
	}
	return n
}

// stickerReaction is a sticker reacting to the message reactParent, as it
// comes: a sticker message with a STICKER_ANNOTATION association.
func stickerReaction() *waE2E.Message {
	return &waE2E.Message{
		MessageContextInfo: &waE2E.MessageContextInfo{MessageAssociation: &waE2E.MessageAssociation{
			AssociationType:  waE2E.MessageAssociation_STICKER_ANNOTATION.Enum(),
			ParentMessageKey: &waCommon.MessageKey{ID: proto.String(reactParent)},
		}},
		StickerMessage: &waE2E.StickerMessage{URL: proto.String("https://mmg/sticker"), DirectPath: proto.String("/v/sticker"),
			MediaKey: bytes.Repeat([]byte{3}, 32), Mimetype: proto.String("application/was"), IsLottie: proto.Bool(true)},
	}
}

// A sticker reaction shows on the message it reacts to, not as a message,
// in whichever order they come; its sticker downloads under its own ID,
// and deleting it for everyone takes the reaction back.
func TestStickerReaction(t *testing.T) {
	for _, first := range []string{reactParent, reactChild} {
		b := testBackend(t)
		text := &waE2E.Message{Conversation: proto.String("test")}
		if first == reactParent {
			reactArrive(t, b, reactParent, 1000, text)
			reactArrive(t, b, reactChild, 1030, stickerReaction())
		} else {
			reactArrive(t, b, reactChild, 1030, stickerReaction()) // as in history
			reactArrive(t, b, reactParent, 1000, text)
		}
		reactArrive(t, b, reactChild, 1030, stickerReaction()) // again
		if n := reactNewMessages(b); n != 1 {
			t.Errorf("%s first: %d new messages, want 1", first, n)
		}
		ms := b.Messages(reactChat, 10)
		if len(ms) != 1 || ms[0].ID != reactParent || ms[0].Text != "test" {
			t.Fatalf("%s first: messages = %+v, want the one reacted to", first, ms)
		}
		st := ms[0].Stickers
		if len(st) != 1 || st[0].ID != reactChild || !st[0].Lottie || st[0].Me || st[0].SenderID != "9@lid" ||
			!st[0].Time.Equal(time.Unix(1030, 0)) {
			t.Fatalf("%s first: sticker reactions = %+v", first, st)
		}
		media, blob, err := b.store.mediaBlob(b.ctx, reactChat, reactChild)
		var s waE2E.StickerMessage
		if err != nil || media != model.MediaSticker || proto.Unmarshal(blob, &s) != nil || s.GetDirectPath() != "/v/sticker" {
			t.Fatalf("%s first: the sticker's media = %v, %v, %v", first, media, s.GetDirectPath(), err)
		}

		// The same person's next sticker takes the first one's place.
		reactArrive(t, b, "NEXT", 1040, stickerReaction())
		if st := b.Messages(reactChat, 10)[0].Stickers; len(st) != 1 || st[0].ID != "NEXT" {
			t.Fatalf("%s first: after another sticker = %+v", first, st)
		}

		for _, id := range []string{"NEXT", reactChild} {
			b.onMessage(&events.Message{
				Info: types.MessageInfo{MessageSource: types.MessageSource{Chat: types.NewJID("123", types.GroupServer),
					Sender: types.NewJID("9", types.HiddenUserServer), IsGroup: true}, ID: "R" + id, Timestamp: time.Unix(1050, 0)},
				Message: &waE2E.Message{ProtocolMessage: &waE2E.ProtocolMessage{
					Key: &waCommon.MessageKey{ID: proto.String(id)}, Type: waE2E.ProtocolMessage_REVOKE.Enum()}},
			})
		}
		ms = b.Messages(reactChat, 10)
		if len(ms) != 1 || len(ms[0].Stickers) != 0 || ms[0].Kind == model.KindDeleted {
			t.Fatalf("%s first: after taking it back = %+v", first, ms)
		}
	}
}

// A sticker reaction an older version stored as a message of the chat is
// filed under the message it reacts to.
func TestFoldStickerReacts(t *testing.T) {
	b := testBackend(t)
	reactArrive(t, b, reactParent, 1000, &waE2E.Message{Conversation: proto.String("test")})
	sticker := &model.Message{ChatID: reactChat, ID: reactChild, Kind: model.KindSticker, Media: model.MediaSticker, Time: time.Unix(1030, 0)}
	if err := b.store.putMessage(b.ctx, b.db, storedMsg{Message: sticker, senderJID: "9@lid", rawPayload: marshal(stickerReaction())}); err != nil {
		t.Fatal(err)
	}
	if n := len(b.Messages(reactChat, 10)); n != 2 {
		t.Fatalf("%d messages before, want 2", n)
	}
	if _, err := b.db.Exec(`DELETE FROM wz_meta WHERE key = 'sticker_reacts_folded'`); err != nil {
		t.Fatal(err)
	}
	if err := b.store.foldStickerReacts(b.ctx); err != nil {
		t.Fatal(err)
	}
	ms := b.Messages(reactChat, 10)
	if len(ms) != 1 || len(ms[0].Stickers) != 1 || ms[0].Stickers[0].ID != reactChild || ms[0].Stickers[0].SenderID != "9@lid" {
		t.Fatalf("after folding = %+v", ms)
	}
}
