// Package sound plays the notification sounds the app offers, so a pick
// from Settings > Notifications can be heard before it's chosen.
//
// Windows' toast can only ask for the system's own sounds: it takes an
// "ms-winsoundevent:" name and won't read a file of ours. So the sounds are
// played by the app itself, from files it writes out of its own build.
//
// Each system has its own file (player_windows.go, ...). Where there is
// none, Play does nothing.
package sound

import (
	"crypto/sha1"
	"embed"
	"encoding/hex"
	"os"
	"path/filepath"
)

// Choice is a notification sound a user can pick.
type Choice struct {
	Key  string // the preference's value, such as "1"
	Name string // as Settings lists it
	file string // name in the embedded sounds, when it has one
}

// Choices are the sounds Settings offers, in order. The first is the
// system's own notification sound, which the toast plays itself.
var Choices = []Choice{
	{Key: "", Name: "Default"},
	{Key: "1", Name: "Sound 1", file: "sound-1.mp3"},
	{Key: "2", Name: "Sound 2", file: "sound-2.mp3"},
	{Key: "3", Name: "Sound 3", file: "sound-3.mp3"},
}

//go:embed sounds/*.mp3
var files embed.FS

// Path returns the file holding the sound key picks, writing it to dir
// unless it's there already: the system's players read files, not memory.
// It returns "" for the default sound, which the toast plays.
func Path(dir, key string) string {
	f := fileFor(key)
	if f == "" || dir == "" {
		return ""
	}
	data, err := files.ReadFile("sounds/" + f)
	if err != nil {
		return ""
	}
	sum := sha1.Sum(data)
	path := filepath.Join(dir, hex.EncodeToString(sum[:10])+".mp3")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return ""
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return ""
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return ""
	}
	return path
}

// Name is how key's sound is called, for a notice.
func Name(key string) string {
	for _, c := range Choices {
		if c.Key == key {
			return c.Name
		}
	}
	return Choices[0].Name
}

func fileFor(key string) string {
	for _, c := range Choices {
		if c.Key == key {
			return c.file
		}
	}
	return ""
}
