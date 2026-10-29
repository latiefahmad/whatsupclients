//go:build windows

package notify

import (
	"strings"
	"testing"
	"time"
)

// TestToastAudioBeforeActions checks that <audio> comes before <actions>:
// the toast schema requires it there, and Windows silently drops a toast's
// sound when the element is anywhere else.
func TestToastAudioBeforeActions(t *testing.T) {
	x := toastXML(Notification{
		ID:    "ana@lid",
		Title: "Ana",
		Body:  "hi",
		Time:  time.Unix(1_700_000_000, 0),
		Reply: true,
	}, "")
	audio := strings.Index(x, "<audio ")
	actions := strings.Index(x, "<actions>")
	visual := strings.Index(x, "</visual>")
	if audio < 0 || actions < 0 || visual < 0 {
		t.Fatalf("toast is missing a part: %s", x)
	}
	if !(visual < audio && audio < actions) {
		t.Errorf("want visual < audio < actions, got %d, %d, %d: %s", visual, audio, actions, x)
	}
	if !strings.Contains(x, `src="ms-winsoundevent:Notification.IM"`) {
		t.Errorf("no sound requested: %s", x)
	}
}

// TestToastSilentAudio checks that a silent toast says so in the same place,
// which is what keeps it quiet.
func TestToastSilentAudio(t *testing.T) {
	x := toastXML(Notification{ID: "ana@lid", Title: "Ana", Body: "hi", Reply: true, Silent: true}, "")
	if !strings.Contains(x, `<audio silent="true"/>`) {
		t.Errorf("not silent: %s", x)
	}
	if strings.Index(x, "<audio ") > strings.Index(x, "<actions>") {
		t.Errorf("silent audio after the actions: %s", x)
	}
}
