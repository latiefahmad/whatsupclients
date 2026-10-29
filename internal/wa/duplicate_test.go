package wa

import (
	"testing"
	"time"

	"github.com/polymorfa/hypermeow/proto/waCommon"
	"github.com/polymorfa/hypermeow/proto/waE2E"
	"github.com/polymorfa/hypermeow/types"
	"github.com/polymorfa/hypermeow/types/events"
	"google.golang.org/protobuf/proto"
)

// TestHDPairIsOneMessage checks that the two messages WhatsApp sends for one
// photo picked as HD (the original and its HD_IMAGE_DUAL_UPLOAD copy) show
// as one, not as the same picture twice.
func TestHDPairIsOneMessage(t *testing.T) {
	b := testBackend(t)
	db := b.db

	chat := types.NewJID("628123", types.DefaultUserServer)
	me := types.NewJID("628999", types.DefaultUserServer)
	at := time.Unix(1700000000, 0)

	photo := func(id string, hd bool) *events.Message {
		img := &waE2E.ImageMessage{
			Caption:       proto.String("foto lawas"),
			FileLength:    proto.Uint64(1234),
			Mimetype:      proto.String("image/jpeg"),
			URL:           proto.String("https://mmg.whatsapp.net/x"),
			MediaKey:      make([]byte, 32),
			FileSHA256:    make([]byte, 32),
			FileEncSHA256: make([]byte, 32),
			DirectPath:    proto.String("/v/x"),
		}
		msg := &waE2E.Message{ImageMessage: img}
		if hd {
			msg.MessageContextInfo = &waE2E.MessageContextInfo{
				MessageAssociation: &waE2E.MessageAssociation{
					AssociationType: waE2E.MessageAssociation_HD_IMAGE_DUAL_UPLOAD.Enum(),
					ParentMessageKey: &waCommon.MessageKey{
						ID: proto.String("ORIGINAL1"),
					},
				},
			}
		}
		return &events.Message{
			Info: types.MessageInfo{
				ID:        id,
				Timestamp: at,
				MessageSource: types.MessageSource{
					Chat: chat, Sender: me, IsFromMe: true,
				},
			},
			Message:    msg,
			RawMessage: msg,
		}
	}

	b.onMessage(photo("ORIGINAL1", false))
	b.onMessage(photo("HDCOPY1", true))

	var n int
	if err := db.QueryRow(`SELECT count(*) FROM wz_messages WHERE chat = ?`, chat.String()).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("one photo sent as HD is stored %d times, want 1", n)
	}
}
