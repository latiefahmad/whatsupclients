package wa

import (
	"testing"

	"github.com/polymorfa/hypermeow/proto/waAICommonDeprecated"
	"github.com/polymorfa/hypermeow/proto/waE2E"
	"google.golang.org/protobuf/proto"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

func richMsg(parts ...string) *waE2E.Message {
	e := &waE2E.AIRichResponseMessage{}
	for _, p := range parts {
		e.Submessages = append(e.Submessages, &waAICommonDeprecated.AIRichResponseSubMessage{
			MessageType: waAICommonDeprecated.AIRichResponseSubMessageType_AI_RICH_RESPONSE_TEXT.Enum(),
			MessageText: proto.String(p),
		})
	}
	return &waE2E.Message{RichResponseMessage: e}
}

// TestRichResponseShowsText checks that an AI agent's reply is shown, not left
// as "This message couldn't load": rich responses had no case in describe, so
// they fell through to KindUnsupported.
func TestRichResponseShowsText(t *testing.T) {
	got := describe(richMsg("Halo! Ada yang bisa saya bantu?", "Notchy tadi sudah aktif."))
	if got.kind == model.KindUnsupported {
		t.Fatal("rich response became KindUnsupported")
	}
	want := "Halo! Ada yang bisa saya bantu?\n\nNotchy tadi sudah aktif."
	if got.text != want {
		t.Errorf("text: %q, want %q", got.text, want)
	}
}

// TestRichResponseSkipsEmptyParts checks that blank submessages, which the
// agent sends between sections, don't leave gaps in the text.
func TestRichResponseSkipsEmptyParts(t *testing.T) {
	got := describe(richMsg("satu", "  ", "", "dua"))
	if got.text != "satu\n\ndua" {
		t.Errorf("text: %q, want %q", got.text, "satu\n\ndua")
	}
}

// TestRichResponseWithoutText checks that a rich response carrying only
// non-text parts (a table, an image) still reads as unsupported rather than
// as an empty message: there is nothing to show, but the message exists.
func TestRichResponseWithoutText(t *testing.T) {
	e := &waE2E.AIRichResponseMessage{Submessages: []*waAICommonDeprecated.AIRichResponseSubMessage{{
		MessageType: waAICommonDeprecated.AIRichResponseSubMessageType_AI_RICH_RESPONSE_GRID_IMAGE.Enum(),
	}}}
	got := describe(&waE2E.Message{RichResponseMessage: e})
	if got.text != "" {
		t.Errorf("text: %q, want empty", got.text)
	}
	if got.kind != model.KindText {
		t.Errorf("kind: %v, want KindText so it isn't shown as unreadable", got.kind)
	}
}

// TestRichResponseIsNotUnsupported checks the whole path: hasContent must see
// the text, so parse doesn't mark the message unreadable.
func TestRichResponseIsNotUnsupported(t *testing.T) {
	m := richMsg("jawaban")
	if !hasContent(m) {
		t.Error("hasContent: false, want true for a rich response with text")
	}
}
