//go:build windows

package sound

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/latiefahmad/whatsupclients/internal/video"
)

var errPlaySound = errors.New("sound: PlaySound failed")

// The system's own player decodes and outputs the file, so the app needs no
// decoder of its own (see internal/video). The default sound is Windows' own,
// which winmm plays from its alias.

func init() {
	opener = open
	demo = playDefault
}

func open(path string) (Player, error) {
	p, err := video.OpenAudio(path, nil)
	if err != nil {
		return nil, err
	}
	p.Play()
	return &audioPlayer{p: p}, nil
}

type audioPlayer struct{ p *video.Player }

func (a *audioPlayer) Close() { a.p.Close() }

var (
	winmm          = syscall.NewLazyDLL("winmm.dll")
	procPlaySound  = winmm.NewProc("PlaySoundW")
	soundAsync     = uintptr(0x0001)
	soundNoDefault = uintptr(0x0002)
	soundAlias     = "SystemNotification"
)

func playDefault() (Player, error) {
	if err := winmm.Load(); err != nil {
		return nil, err
	}
	name, err := syscall.UTF16PtrFromString(soundAlias)
	if err != nil {
		return nil, err
	}
	if r, _, _ := procPlaySound.Call(uintptr(unsafe.Pointer(name)), 0, soundAsync|soundNoDefault); r == 0 {
		return nil, errPlaySound
	}
	return defaultSound{}, nil
}

// defaultSound is winmm's own playback, which stops with another alias.
type defaultSound struct{}

func (defaultSound) Close() {
	procPlaySound.Call(0, 0, 0)
}
