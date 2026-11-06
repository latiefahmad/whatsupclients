package wa

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	waServerSync "github.com/polymorfa/hypermeow/proto/waServerSync"
)

// snapshotBlob returns a snapshot's blob, downloading it at most once per
// handle. Fetching needs the media connection, which is slow to refresh and
// drops with the socket, so a sync attempt that fails there would otherwise
// never get to the analysis that explains the refusal. The bytes are kept in
// the data directory, named by handle, and read from there afterwards.
func (b *Backend) snapshotBlob(ref *waServerSync.ExternalBlobReference) ([]byte, error) {
	handle := ref.GetHandle()
	if handle == "" {
		return nil, errors.New("the reference has no handle")
	}
	path := filepath.Join(b.dataDir, "appstate-snapshots", shortBytes([]byte(handle)))
	if raw, err := os.ReadFile(path); err == nil {
		return raw, nil
	}

	cli := b.client()
	if cli == nil {
		return nil, errors.New("not connected")
	}
	ctx, cancel := context.WithTimeout(b.ctx, 60*time.Second)
	defer cancel()
	raw, err := cli.Download(ctx, ref)
	if err != nil {
		return nil, err
	}
	if want := int(ref.GetFileSizeBytes()); want > 0 && len(raw) != want {
		return nil, fmt.Errorf("got %d of %d bytes", len(raw), want)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return raw, nil
	}
	// A failed write only costs another download, so the analysis still runs.
	_ = os.WriteFile(path, raw, 0o600)
	return raw, nil
}
