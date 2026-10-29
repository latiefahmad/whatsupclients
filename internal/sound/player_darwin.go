//go:build darwin

package sound

import (
	"os/exec"
	"sync"
)

// afplay is macOS's own player, which decodes the files as they are. The
// default sound is the system's Glass.aiff.

func init() {
	opener = open
	demo = playDefault
}

const defaultSound = "/System/Library/Sounds/Glass.aiff"

func open(path string) (Player, error) {
	cmd := exec.Command("afplay", path)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	p := &afPlayer{cmd: cmd}
	go cmd.Wait()
	return p, nil
}

func playDefault() (Player, error) { return open(defaultSound) }

type afPlayer struct {
	cmd  *exec.Cmd
	once sync.Once
}

func (a *afPlayer) Close() {
	a.once.Do(func() {
		if a.cmd.Process != nil {
			a.cmd.Process.Kill()
		}
	})
}
