package ui

import (
	"testing"
	"time"
)

// TestResting: animations rest while the user is idle, and play again
// once they're back at the keyboard.
func TestResting(t *testing.T) {
	st := newSlashTest(t, "work")
	u := st.u
	old := idleTime
	idleTime = func() time.Duration { return idleAfter + time.Minute }
	defer func() {
		u.stopResting() // end the watcher before the stub goes back
		idleTime = old
	}()

	// updateResting looks at the idle time at most once a second.
	st.now = st.now.Add(restCheck)
	st.frame()
	if !u.rest.on || u.rest.watch == nil {
		t.Fatal("animations don't rest while the user is idle")
	}
	u.animFrame("rest/test", 64, false, func() []byte { return nil })
	if len(u.players.m) != 0 {
		t.Error("an animated sticker started while the user is idle")
	}

	u.stopResting()
	idleTime = old
	st.now = st.now.Add(restCheck)
	st.frame()
	if u.rest.on || u.rest.watch != nil {
		t.Fatal("animations don't come back once the user is back")
	}
	n := len(u.players.m) // the chat's own players are back on screen
	u.animFrame("rest/test", 64, false, func() []byte { return nil })
	if len(u.players.m) != n+1 {
		t.Error("an animated sticker didn't start once the user is back")
	}
}
