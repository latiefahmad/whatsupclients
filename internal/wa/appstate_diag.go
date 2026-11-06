package wa

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/polymorfa/hypermeow/appstate"
	"github.com/polymorfa/hypermeow/appstate/lthash"
	waBinary "github.com/polymorfa/hypermeow/binary"
	waServerSync "github.com/polymorfa/hypermeow/proto/waServerSync"
	"github.com/polymorfa/hypermeow/util/hkdfutil"
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
	var snapRef *waServerSync.ExternalBlobReference
	if hasSnap {
		var ref waServerSync.ExternalBlobReference
		if err := proto.Unmarshal(raw, &ref); err != nil {
			b.log.Warnf("app state diag: collection %s: snapshot reference unreadable: %v", name, err)
		} else {
			snapRef = &ref
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
	b.inspectSnapshotBlob(name, snapRef, patchNodes)
}

// inspectSnapshotBlob downloads the snapshot's blob and counts what it holds.
// A snapshot carries records only, all of them SET, so updateHash adds one
// value MAC per record and subtracts nothing: the hash is decided entirely by
// these records and the version. The count and version tell a truncated
// download (fewer entries than the MAC was made from) from a mismatched key.
func (b *Backend) inspectSnapshotBlob(name appstate.WAPatchName, ref *waServerSync.ExternalBlobReference, patchNodes []waBinary.Node) {
	if ref == nil {
		return
	}
	cli := b.client()
	if cli == nil {
		return
	}
	ctx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
	defer cancel()
	raw, err := cli.Download(ctx, ref)
	if err != nil {
		b.log.Warnf("app state diag: snapshot %s: blob download failed: %v", name, err)
		return
	}
	var ss waServerSync.SyncdSnapshot
	if err := proto.Unmarshal(raw, &ss); err != nil {
		b.log.Warnf("app state diag: snapshot %s: blob has %d bytes, unreadable: %v", name, len(raw), err)
		return
	}
	records := ss.GetRecords()
	shortValues := 0
	for _, rec := range records {
		if len(rec.GetValue().GetBlob()) < 32 {
			shortValues++
		}
	}
	b.log.Warnf("app state diag: snapshot %s: blob %d bytes, v%d records=%d short_values=%d keyid=%s expected_filesize=%d",
		name, len(raw), ss.GetVersion().GetVersion(), len(records), shortValues,
		shortBytes(ss.GetKeyID().GetID()), ref.GetFileSizeBytes())

	// The decisive number: the snapshot carries its own MAC, so recomputing
	// the hash over exactly its records says whether the mismatch comes from
	// the key or from the hash hypermeow builds. The value MAC of a SET is the
	// last 32 bytes of its blob, and a snapshot subtracts nothing.
	key, err := cli.Store.AppStateKeys.GetAppStateSyncKey(ctx, ss.GetKeyID().GetID())
	if err != nil || key == nil {
		b.log.Warnf("app state diag: snapshot %s: key %s not held: %v", name, shortBytes(ss.GetKeyID().GetID()), err)
		return
	}
	var added [][]byte
	for _, rec := range records {
		if value := rec.GetValue().GetBlob(); len(value) >= 32 {
			added = append(added, value[len(value)-32:])
		}
	}
	var hash [128]byte
	lthash.WAPatchIntegrity.SubtractThenAddInPlace(hash[:], nil, added)

	expanded := expandKey(key.Data)
	want := snapshotMAC(hash, ss.GetVersion().GetVersion(), name, expanded)
	b.log.Warnf("app state diag: snapshot %s: hash=%s v%d, computed mac=%s snapshot mac=%s match=%v",
		name, shortHash(hash), ss.GetVersion().GetVersion(), shortBytes(want), shortBytes(ss.GetMac()),
		hmac.Equal(want, ss.GetMac()))

	// If none of the five subkeys reproduces the MAC, the hash itself is wrong
	// and no key can explain the mismatch. Trying each one is the only way to
	// say that without assuming which subkey the server signed with.
	for _, sub := range []struct {
		name string
		at   int
	}{
		{"index", 0}, {"value_encryption", 32}, {"value_mac", 64}, {"snapshot_mac", 96}, {"patch_mac", 128},
	} {
		got := snapshotMACKeyed(hash, ss.GetVersion().GetVersion(), name, expanded[sub.at:sub.at+32])
		if hmac.Equal(got, ss.GetMac()) {
			b.log.Warnf("app state diag: snapshot %s: the MAC matches the %s subkey", name, sub.name)
		}
	}

	// A snapshot's version is what it was sealed at, and the MAC covers it, so
	// a MAC made at another version can't verify. If one of these nearby
	// versions matches, the records are fine and only the version is off.
	for _, v := range []uint64{ss.GetVersion().GetVersion() - 1, ss.GetVersion().GetVersion() + 1} {
		if hmac.Equal(snapshotMACKeyed(hash, v, name, expanded[96:128]), ss.GetMac()) {
			b.log.Warnf("app state diag: snapshot %s: the MAC matches at v%d, not v%d",
				name, v, ss.GetVersion().GetVersion())
		}
	}

	// The same MAC with the version taken from the patches rather than the
	// snapshot, in case the two disagree about which version was sealed.
	if len(patchNodes) > 0 {
		if raw, ok := patchNodes[len(patchNodes)-1].Content.([]byte); ok {
			var last waServerSync.SyncdPatch
			if proto.Unmarshal(raw, &last) == nil {
				b.log.Warnf("app state diag: snapshot %s: newest patch is v%d (snapshot claims v%d)",
					name, last.GetVersion().GetVersion(), ss.GetVersion().GetVersion())
			}
		}
	}
}

// snapshotMACKeyed signs a state hash with one specific subkey.
func snapshotMACKeyed(hash [128]byte, version uint64, name appstate.WAPatchName, key []byte) []byte {
	m := hmac.New(sha256.New, key)
	m.Write(hash[:])
	var v [8]byte
	binary.BigEndian.PutUint64(v[:], version)
	m.Write(v[:])
	m.Write([]byte(name))
	return m.Sum(nil)
}

// expandKey turns stored key data into the five subkeys a collection uses, the
// way hypermeow does ("WhatsApp Mutation Keys", 160 bytes).
func expandKey(data []byte) []byte {
	return hkdfutil.SHA256(data, nil, []byte("WhatsApp Mutation Keys"), 160)
}

// snapshotMAC is the HMAC a snapshot's MAC is checked against: the state hash,
// the version, then the collection's name, under the snapshot key.
func snapshotMAC(hash [128]byte, version uint64, name appstate.WAPatchName, expanded []byte) []byte {
	return snapshotMACKeyed(hash, version, name, expanded[96:128])
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
