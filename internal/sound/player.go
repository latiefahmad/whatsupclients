package sound

import (
	"sync"
	"time"
)

// Player plays one sound file. A system backend implements the interface
// its own player provides; this package keeps at most one going, so a
// burst of messages plays one sound rather than a queue of them.
type Player interface {
	// Close stops playback and frees it.
	Close()
}

// opener opens a file for playback; a system backend sets it. demo plays
// the system's own notification sound.
var opener func(path string) (Player, error)

var demo func() (Player, error)

// Init sets the directory the sound files are written to (the same one the
// notifications keep their pictures in). It's called once, at startup.
func Init(dir string) {
	mu.Lock()
	keep = dir
	mu.Unlock()
}

// Play plays the sound key picks. It returns whether anything was played:
// the default sound is the toast's own, and a system without a player does
// nothing.
//
// A sound started less than minGap ago is not played again, so a crowd of
// messages makes one noise.
func Play(key string) bool {
	mu.Lock()
	dir := keep
	mu.Unlock()
	path := Path(dir, key)
	if path == "" || opener == nil {
		return false
	}
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	if !last.IsZero() && now.Sub(last) < minGap {
		return false
	}
	last = now
	if active != nil {
		active.Close()
		active = nil
	}
	p, err := opener(path)
	if err != nil {
		return false
	}
	active = p
	return true
}

// Preview plays the system's own notification sound, so Settings can let
// its Default choice be heard too.
func Preview() {
	if demo == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if active != nil {
		active.Close()
		active = nil
	}
	if p, err := demo(); err == nil {
		active = p
	}
}

// StopSilently ends a sound still playing (the chat was read).
func StopSilently() {
	mu.Lock()
	defer mu.Unlock()
	if active != nil {
		active.Close()
		active = nil
	}
}

// minGap is how long after one sound the next may play.
const minGap = 1200 * time.Millisecond

var last time.Time

// mu guards active and last.
var mu sync.Mutex

// active is the sound playing, if any; keep is where files are written.
var (
	active Player
	keep   string
)
