package wa

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	waServerSync "github.com/polymorfa/hypermeow/proto/waServerSync"
	waLog "github.com/polymorfa/hypermeow/util/log"
	"google.golang.org/protobuf/proto"
)

func snapshotRef(t *testing.T, handle string, size uint64) *waServerSync.ExternalBlobReference {
	t.Helper()
	return &waServerSync.ExternalBlobReference{Handle: &handle, FileSizeBytes: &size}
}

// TestSnapshotBlobReadFromDisk checks the point of the cache: once the blob is
// on disk, no client is needed, so a sync attempt that can't reach the media
// server still gets its analysis.
func TestSnapshotBlobReadFromDisk(t *testing.T) {
	dir := t.TempDir()
	want := bytes.Repeat([]byte{9}, 1000)
	handle := "4:abc123"

	path := filepath.Join(dir, "appstate-snapshots", shortBytes([]byte(handle)))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatal(err)
	}

	// No client at all: reading the cache must not need one.
	b := &Backend{dataDir: dir, log: waLog.Noop}
	got, err := b.snapshotBlob(snapshotRef(t, handle, 1000))
	if err != nil {
		t.Fatalf("reading the cached blob failed: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %d bytes, want %d", len(got), len(want))
	}
}

// TestSnapshotBlobWithoutClientOrCacheIsAnError keeps a missing blob from
// being reported as an empty one, which would read as a truncated snapshot.
func TestSnapshotBlobWithoutClientOrCacheIsAnError(t *testing.T) {
	b := &Backend{dataDir: t.TempDir(), log: waLog.Noop}
	if _, err := b.snapshotBlob(snapshotRef(t, "4:nocache", 500)); err == nil {
		t.Fatal("a blob that is neither cached nor downloadable returned no error")
	}
}

// TestSnapshotBlobRejectsAHandlelessRef avoids writing a file named after an
// empty handle, which every handleless reference would share.
func TestSnapshotBlobRejectsAHandlelessRef(t *testing.T) {
	b := &Backend{dataDir: t.TempDir(), log: waLog.Noop}
	if _, err := b.snapshotBlob(&waServerSync.ExternalBlobReference{}); err == nil {
		t.Fatal("a reference with no handle was accepted")
	}
}

// TestSnapshotBlobCacheRoundTrip checks the size check isn't hit on a good
// read and that the bytes survive a write.
func TestSnapshotBlobCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	handle := "4:roundtrip"
	ss := &waServerSync.SyncdSnapshot{}
	raw, err := proto.Marshal(ss)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "appstate-snapshots", shortBytes([]byte(handle)))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	b := &Backend{dataDir: dir, log: waLog.Noop}
	got, err := b.snapshotBlob(snapshotRef(t, handle, uint64(len(raw))))
	if err != nil {
		t.Fatal(err)
	}
	var back waServerSync.SyncdSnapshot
	if err := proto.Unmarshal(got, &back); err != nil {
		t.Fatalf("the cached bytes don't parse: %v", err)
	}
}
