package wa

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/polymorfa/hypermeow/proto/waAICommon"
	"github.com/polymorfa/hypermeow/proto/waE2E"
	"google.golang.org/protobuf/proto"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

// TestRereadRichResponses checks that a bot reply stored as unreadable before
// the app could read one gets its text back, from the payload it kept.
func TestRereadRichResponses(t *testing.T) {
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

	raw, err := proto.Marshal(&waE2E.Message{RichResponseMessage: &waE2E.AIRichResponseMessage{
		UnifiedResponse: &waAICommon.AIRichResponseUnifiedResponse{
			Data: []byte(`{"sections":[{"view_model":{"primitive":{"text":"Halo dari bot"}}}]}`),
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO wz_messages (chat, id, ts, kind, text, raw_payload)
		VALUES ('b@bot', 'm1', 1, ?, '', ?)`, int(model.KindUnsupported), raw); err != nil {
		t.Fatal(err)
	}
	// init ran the reread already, on an empty table.
	if _, err := db.ExecContext(ctx, `DELETE FROM wz_meta WHERE key = 'rich_responses_reread'`); err != nil {
		t.Fatal(err)
	}

	if err := s.rereadRichResponses(ctx); err != nil {
		t.Fatal(err)
	}

	var kind, text string
	if err := db.QueryRowContext(ctx, `SELECT kind, text FROM wz_messages WHERE id = 'm1'`).
		Scan(&kind, &text); err != nil {
		t.Fatal(err)
	}
	if kind != "0" {
		t.Errorf("kind: %s, want %d (KindText)", kind, model.KindText)
	}
	if text != "Halo dari bot" {
		t.Errorf("text: %q", text)
	}
}

// TestRereadRichResponsesOnce checks that the reread runs once: a message that
// is still unreadable afterwards isn't retried on every start.
func TestRereadRichResponsesOnce(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `INSERT INTO wz_messages (chat, id, ts, kind)
		VALUES ('c', 'm1', 1, ?)`, int(model.KindUnsupported)); err != nil {
		t.Fatal(err)
	}
	if err := s.rereadRichResponses(ctx); err != nil {
		t.Fatal(err)
	}
	if s.meta(ctx, "rich_responses_reread") == "" {
		t.Error("marker not set, want the reread recorded as done")
	}
}
