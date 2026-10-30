package wa

import (
	"context"
	"database/sql"

	"github.com/polymorfa/hypermeow/proto/waE2E"
	"google.golang.org/protobuf/proto"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

// RereadRichResponsesForTest runs the reread on an opened database, for the
// one-off check of a real store outside the package's tests.
func RereadRichResponsesForTest(db *sql.DB) error {
	return (&msgStore{db: db}).rereadRichResponses(context.Background())
}

// rereadRichResponses gives bot replies that were stored as unreadable their
// text, once. A rich response had no reader, so it was kept as
// KindUnsupported with no text; the payload was stored all the same, and it
// reads now.
func (s *msgStore) rereadRichResponses(ctx context.Context) error {
	const done = "rich_responses_reread"
	if s.meta(ctx, done) != "" {
		return nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT rowid, raw_payload FROM wz_messages
		WHERE kind = ?1 AND length(raw_payload) > 0`, int(model.KindUnsupported))
	if err != nil {
		return err
	}
	type fix struct {
		rowid int64
		text  string
	}
	var fixes []fix
	for rows.Next() {
		var id int64
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			rows.Close()
			return err
		}
		m := &waE2E.Message{}
		if proto.Unmarshal(raw, m) != nil {
			continue
		}
		r := m.GetRichResponseMessage()
		if r == nil {
			continue
		}
		if t := richResponse(r).text; t != "" {
			fixes = append(fixes, fix{rowid: id, text: t})
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, f := range fixes {
		if _, err := tx.ExecContext(ctx,
			`UPDATE wz_messages SET kind = ?1, text = ?2 WHERE rowid = ?3`,
			int(model.KindText), f.text, f.rowid); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO wz_meta (key, value) VALUES (?1, '1')
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, done); err != nil {
		return err
	}
	return tx.Commit()
}
