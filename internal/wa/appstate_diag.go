package wa

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/polymorfa/hypermeow/appstate"
	waBinary "github.com/polymorfa/hypermeow/binary"
	waServerSync "github.com/polymorfa/hypermeow/proto/waServerSync"
	"google.golang.org/protobuf/proto"
)

// Diagnostics for app state hash breakage. A collection whose LTHash stops
// verifying can't send or receive changes, and this device and the phone
// drift apart. The logs showed it healing from the phone's copy and breaking
// again within a minute, but not what makes it break, so these lines record
// enough to tell the causes apart: a patch the server answered 409 to (its
// conflicting patches apply to the state read before the send), a version
// that jumps without the hash following, and the phone's copy being refused
// for being older than this device.
//
// They are logged only with -debug, and never change behaviour.

// shortBytes names a short blob by its first and last bytes.
func shortBytes(v []byte) string {
	if len(v) == 0 {
		return "(none)"
	}
	if len(v) <= 8 {
		return hex.EncodeToString(v)
	}
	return hex.EncodeToString(v[:4]) + ".." + hex.EncodeToString(v[len(v)-4:])
}

// noteAppStateSnapshot reads the shape of an app state collection the server
// answered with. It runs on hypermeow's dispatch goroutine, so it only parses
// and logs: the node goes on untouched and hypermeow still verifies and
// decodes it as usual.
//
// A snapshot travels as an ExternalBlobReference (308 bytes in the case this
// was added for) pointing at a blob fetched separately, so the reference's
// size says nothing about the state it holds. It logs the reference and the
// patches that came with it; the blob itself is left to hypermeow, which
// downloads it as part of the verification this line explains.
func (b *Backend) noteAppStateSnapshot(_ context.Context, node *waBinary.Node) {
	if !b.debugLog || node.Tag != "iq" {
		return
	}
	collection, ok := node.GetOptionalChildByTag("sync", "collection")
	if !ok {
		return
	}
	name := appstate.WAPatchName(collection.AttrGetter().OptionalString("name"))
	snapNode := collection.GetChildByTag("snapshot")
	raw, hasSnap := snapNode.Content.([]byte)
	if hasSnap {
		var ref waServerSync.ExternalBlobReference
		if err := proto.Unmarshal(raw, &ref); err != nil {
			b.log.Warnf("app state diag: collection %s: snapshot reference unreadable: %v", name, err)
		} else {
			b.log.Warnf("app state diag: snapshot %s: ref bytes=%d handle=%s filesize=%d sha256=%s",
				name, len(raw), ref.GetHandle(), ref.GetFileSizeBytes(), shortBytes(ref.GetFileSHA256()))
		}
	} else {
		b.log.Warnf("app state diag: collection %s: no snapshot at all", name)
	}
	patchesNode := collection.GetChildByTag("patches")
	patchNodes := patchesNode.GetChildren()
	for i, pn := range patchNodes {
		raw, ok := pn.Content.([]byte)
		if !ok || pn.Tag != "patch" {
			continue
		}
		var p waServerSync.SyncdPatch
		if err := proto.Unmarshal(raw, &p); err != nil {
			continue
		}
		if i < 8 {
			b.log.Warnf("app state diag: snapshot %s: patch v%d mutations=%d external=%v",
				name, p.GetVersion().GetVersion(), len(p.GetMutations()), p.GetExternalMutations() != nil)
		}
	}
	b.log.Infof("app state diag: collection %s: %d patch(es)", name, len(patchNodes))
}

// appStateStamp reads the local version and hash of a collection.
func (b *Backend) appStateStamp(name appstate.WAPatchName) string {
	cli := b.client()
	if cli == nil {
		return "no client"
	}
	v, hash, err := cli.Store.AppState.GetAppStateVersion(b.ctx, string(name))
	if err != nil {
		return fmt.Sprintf("version unreadable: %v", err)
	}
	return fmt.Sprintf("v%d hash=%s", v, shortHash(hash))
}

// shortHash names a hash by its first and last bytes, enough to tell two
// states apart in a log without a page of hex.
func shortHash(h [128]byte) string {
	return hex.EncodeToString(h[:4]) + ".." + hex.EncodeToString(h[124:])
}

// noteAppStateSend records the state a patch is sent from, and whether the
// patch changes it. A 409 means the server's conflicting patches are applied
// to this version; if that apply fails partway, some mutations are stored and
// others aren't, and the hash drifts.
func (b *Backend) noteAppStateSend(name appstate.WAPatchName, mutations int) {
	if !b.debugLog {
		return
	}
	b.log.Debugf("app state diag: sending patch, %s, %d mutation(s): %s",
		name, mutations, b.appStateStamp(name))
}

// noteAppStateResult records the state after a send returned, whatever its
// outcome. A version that went up while the send failed is the drift.
func (b *Backend) noteAppStateResult(name appstate.WAPatchName, before string, sendErr error) {
	if !b.debugLog {
		return
	}
	after := b.appStateStamp(name)
	if sendErr == nil {
		b.log.Debugf("app state diag: sent ok, %s: before %s -> after %s", name, before, after)
		return
	}
	b.log.Warnf("app state diag: send failed, %s: before %s -> after %s, err %v", name, before, after, sendErr)
}

// noteAppStateSyncError records what the collection looked like when a sync
// failed to verify, which is when the hash first shows up broken.
func (b *Backend) noteAppStateSyncError(name appstate.WAPatchName, err error) {
	if !b.debugLog {
		return
	}
	b.log.Warnf("app state diag: sync error %s: %v, local %s", name, err, b.appStateStamp(name))
}

// noteRecoveryRefused records what a request for the phone's copy will meet:
// hypermeow drops that copy when this device's version is already as high,
// and the broken local state inflates its version by counting the patches the
// server refused. So a broken hash can keep refusing its own remedy.
func (b *Backend) noteRecoveryRefused(name appstate.WAPatchName) {
	if !b.debugLog {
		return
	}
	b.log.Infof("app state diag: asking for a repair copy of %s, local %s (hypermeow drops one at or below this version)",
		name, b.appStateStamp(name))
}
