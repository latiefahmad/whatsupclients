package wa

import (
	"context"
	"strings"

	waLog "github.com/polymorfa/hypermeow/util/log"
)

// mergeSenderLIDChats folds the chats a bot opened under a LID nothing maps
// into the chat that LID's phone number is known by, once.
//
// Those messages were stored twice, once per identity, so each one notified
// twice. parse files them under one chat now (see senderLID); this gives the
// ones already stored the same treatment.
func (s *msgStore) mergeSenderLIDChats(ctx context.Context, log waLog.Logger) error {
	const done = "sender_lid_chats_merged"
	if s.meta(ctx, done) != "" {
		return nil
	}

	// A chat under a LID no mapping knows, whose messages say a phone number
	// sent them, is a duplicate of that number's own chat.
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.jid, (
			SELECT m.sender_jid FROM wz_messages m
			WHERE m.chat = c.jid AND m.sender_jid LIKE '%@s.whatsapp.net'
			ORDER BY m.ts DESC LIMIT 1
		)
		FROM wz_chats c
		WHERE c.jid LIKE '%@lid'
			AND NOT EXISTS (SELECT 1 FROM whatsmeow_lid_map w WHERE w.lid || '@lid' = c.jid)
			AND EXISTS (SELECT 1 FROM wz_messages m WHERE m.chat = c.jid)`)
	if err != nil {
		return err
	}
	type pair struct{ lid, pn string }
	var pairs []pair
	for rows.Next() {
		var lid string
		var pn *string
		if err := rows.Scan(&lid, &pn); err != nil {
			rows.Close()
			return err
		}
		if pn == nil || *pn == "" {
			continue
		}
		pairs = append(pairs, pair{lid: lid, pn: *pn})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, p := range pairs {
		// The number's LID is where its conversation now lives; without it
		// there is no chat to fold into.
		var lid string
		if err := s.db.QueryRowContext(ctx, `SELECT w.lid || '@lid' FROM whatsmeow_lid_map w
			WHERE w.pn = ?`, phoneOf(p.pn)).Scan(&lid); err != nil || lid == "" {
			continue
		}
		if !s.hasMessages(ctx, lid) {
			continue
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err := s.mergeChat(ctx, tx, p.lid, lid); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Infof("merged chat %s into %s", p.lid, lid)
	}
	return s.setMetaValue(ctx, done, "1")
}

// phoneOf cuts a phone-number JID down to the number lid_map stores.
func phoneOf(jid string) string {
	if i := strings.Index(jid, "@"); i > 0 {
		return jid[:i]
	}
	return jid
}

func (s *msgStore) hasMessages(ctx context.Context, chat string) bool {
	var one int
	return s.db.QueryRowContext(ctx, `SELECT 1 FROM wz_messages WHERE chat = ? LIMIT 1`, chat).Scan(&one) == nil
}
