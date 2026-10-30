package wa

import (
	"testing"

	"github.com/polymorfa/hypermeow/proto/waAICommon"
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

// TestRichResponseUnified checks the shape newer bots send: no submessages at
// all, the reply is one JSON document in unifiedResponse. This is the payload
// a real bot sent, trimmed of its UUID.
func TestRichResponseUnified(t *testing.T) {
	e := &waE2E.AIRichResponseMessage{
		MessageType: waAICommonDeprecated.AIRichResponseMessageType_AI_RICH_RESPONSE_TYPE_STANDARD.Enum(),
		UnifiedResponse: &waAICommon.AIRichResponseUnifiedResponse{
			Data: []byte(`{"response_id":"x","sections":[{"view_model":{"__typename":"GenAISingleLayoutViewModel","primitive":{"__typename":"GenAIMarkdownTextUXPrimitive","text":"Bisa banget, Hatz — tapi ada"}}}]}`),
		},
	}
	got := describe(&waE2E.Message{RichResponseMessage: e})
	if got.kind == model.KindUnsupported {
		t.Fatal("unified response became KindUnsupported")
	}
	if got.text != "Bisa banget, Hatz — tapi ada" {
		t.Errorf("text: %q", got.text)
	}
}

// TestRichResponseUnifiedSections checks that a reply split into several
// sections keeps them all, in order.
func TestRichResponseUnifiedSections(t *testing.T) {
	data := `{"sections":[
		{"view_model":{"primitive":{"text":"Pertama"}}},
		{"view_model":{"primitive":{"text":"Kedua"}}}]}`
	if got := unifiedText([]byte(data)); got != "Pertama\n\nKedua" {
		t.Errorf("text: %q, want %q", got, "Pertama\n\nKedua")
	}
}

// TestRichResponseUnifiedItems checks a list view model, whose text sits in
// its items rather than in a single primitive.
func TestRichResponseUnifiedItems(t *testing.T) {
	data := `{"sections":[{"view_model":{"items":[
		{"text":"satu"},{"text":"dua"}]}}]}`
	if got := unifiedText([]byte(data)); got != "satu\ndua" {
		t.Errorf("text: %q, want %q", got, "satu\ndua")
	}
}

// TestRichResponseUnifiedGarbage checks that a malformed document leaves the
// message unreadable rather than crashing the reader.
func TestRichResponseUnifiedGarbage(t *testing.T) {
	if got := unifiedText([]byte("not json")); got != "" {
		t.Errorf("text: %q, want empty", got)
	}
	if got := describe(&waE2E.Message{RichResponseMessage: &waE2E.AIRichResponseMessage{
		UnifiedResponse: &waAICommon.AIRichResponseUnifiedResponse{Data: []byte("{")},
	}}); got.text != "" {
		t.Errorf("text: %q, want empty", got.text)
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
