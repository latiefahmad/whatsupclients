package sound

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPathWritesTheSound checks that a chosen sound is written out of the
// build, once, under a name taken from its contents.
func TestPathWritesTheSound(t *testing.T) {
	dir := t.TempDir()
	path := Path(dir, "2")
	if path == "" {
		t.Fatal("no file for Sound 2")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("not written: %v", err)
	}
	if info.Size() == 0 {
		t.Error("the file is empty")
	}
	if again := Path(dir, "2"); again != path {
		t.Errorf("wrote it twice: %q then %q", path, again)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("%d files written, want 1", len(entries))
	}
}

// TestPathDefault checks that the default sound has no file: the toast
// plays it itself.
func TestPathDefault(t *testing.T) {
	if p := Path(t.TempDir(), ""); p != "" {
		t.Errorf("the default sound has a file: %q", p)
	}
}

// TestChoices checks that every choice but the default has its file in the
// build, so none of them is listed and then silent.
func TestChoices(t *testing.T) {
	for _, c := range Choices[1:] {
		data, err := files.ReadFile("sounds/" + c.file)
		if err != nil || len(data) == 0 {
			t.Errorf("%s (%s) isn't in the build: %v", c.Name, c.file, err)
		}
	}
	if len(Choices) != 4 {
		t.Errorf("%d choices, want 4", len(Choices))
	}
}

// TestName checks the names Settings shows.
func TestName(t *testing.T) {
	for _, c := range Choices {
		if got := Name(c.Key); got != c.Name {
			t.Errorf("Name(%q) = %q, want %q", c.Key, got, c.Name)
		}
	}
	if got := Name("nonsense"); got != Choices[0].Name {
		t.Errorf("Name of an unknown key = %q", got)
	}
}

func TestPathNoDir(t *testing.T) {
	if p := Path("", "1"); p != "" {
		t.Errorf("wrote to an empty dir: %q", p)
	}
	if _, err := os.Stat(filepath.Join("sounds", "sound-1.mp3")); err != nil {
		t.Error("the embedded file is missing")
	}
}
