package wa

import (
	"encoding/json"
	"sync"

	"github.com/polymorfa/hypermeow/appstate"
	"google.golang.org/protobuf/proto"

	waSyncAction "github.com/polymorfa/hypermeow/proto/waSyncAction"
)

// A change made here that the phone refused is kept and sent again, instead
// of being dropped. Before this, a delete was applied to this device's store
// and then its app state patch was sent once: when the patch failed — a
// collection whose hash no longer verified, which stays broken for a while —
// the message stayed deleted here and stayed on the phone, and nothing ever
// sent the change again. The two devices disagreed from then on, whichever
// side the change was made on.
//
// pendingPatch is one mutation that could not be sent yet.
type pendingPatch struct {
	Name    string   `json:"name"`
	Index   []string `json:"index"`
	Version int32    `json:"version"`
	// The value as the protobuf it was built from, so sending it again sends
	// the very same change.
	Value []byte `json:"value,omitempty"`
}

// patchQueue keeps the changes a send refused, oldest first, and sends them
// once their collection syncs again.
type patchQueue struct {
	mu      sync.Mutex
	pending []pendingPatch
}

// enqueuePatch records the mutations of a patch that failed, unless the same
// change is already waiting.
func (b *Backend) enqueuePatch(p appstate.PatchInfo) {
	if len(p.Mutations) == 0 {
		return
	}
	b.patches.mu.Lock()
	for _, m := range p.Mutations {
		dup := false
		for _, have := range b.patches.pending {
			if have.Name == string(p.Type) && sameIndex(have.Index, m.Index) {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		q := pendingPatch{Name: string(p.Type), Index: m.Index, Version: m.Version}
		if m.Value != nil {
			if raw, err := proto.Marshal(m.Value); err == nil {
				q.Value = raw
			}
		}
		b.patches.pending = append(b.patches.pending, q)
	}
	n := len(b.patches.pending)
	b.patches.mu.Unlock()
	b.savePatches()
	b.log.Infof("kept app state change for later: %s (%d waiting)", p.Type, n)
}

// forgetPatch drops the mutations of a patch that reached the phone, so a
// drain or the next connection doesn't send them again.
func (b *Backend) forgetPatch(p appstate.PatchInfo) {
	if len(p.Mutations) == 0 {
		return
	}
	b.patches.mu.Lock()
	changed := false
	var keep []pendingPatch
	for _, have := range b.patches.pending {
		drop := false
		for _, m := range p.Mutations {
			if have.Name == string(p.Type) && sameIndex(have.Index, m.Index) {
				drop = true
				break
			}
		}
		if drop {
			changed = true
			continue
		}
		keep = append(keep, have)
	}
	b.patches.pending = keep
	b.patches.mu.Unlock()
	if changed {
		b.savePatches()
	}
}

// drainPatches sends the changes kept for name again. It runs when that
// collection syncs, which is the moment WhatsApp starts accepting them.
func (b *Backend) drainPatches(name appstate.WAPatchName) {
	b.patches.mu.Lock()
	var send, rest []pendingPatch
	for _, p := range b.patches.pending {
		if p.Name == string(name) {
			send = append(send, p)
		} else {
			rest = append(rest, p)
		}
	}
	b.patches.pending = rest
	b.patches.mu.Unlock()
	if len(send) == 0 {
		return
	}
	b.savePatches()

	muts := make([]appstate.MutationInfo, 0, len(send))
	for _, p := range send {
		m := appstate.MutationInfo{Index: p.Index, Version: p.Version}
		if len(p.Value) > 0 {
			v := &waSyncAction.SyncActionValue{}
			if proto.Unmarshal(p.Value, v) == nil {
				m.Value = v
			}
		}
		muts = append(muts, m)
	}
	b.log.Infof("resending %d app state change(s) for %s", len(muts), name)
	b.sendAppState(appstate.PatchInfo{Type: name, Mutations: muts})
}

// savePatches stores the queue, so a change survives a restart.
func (b *Backend) savePatches() {
	b.patches.mu.Lock()
	data, err := json.Marshal(b.patches.pending)
	empty := len(b.patches.pending) == 0
	b.patches.mu.Unlock()
	if err != nil {
		return
	}
	raw := string(data)
	if empty {
		raw = ""
	}
	_ = b.store.setMetaValue(b.ctx, pendingPatchKey, raw)
}

const pendingPatchKey = "appstate_pending"

// loadPatches reads the queue saved by an earlier run.
func (b *Backend) loadPatches() {
	raw := b.store.meta(b.ctx, pendingPatchKey)
	if raw == "" {
		return
	}
	var p []pendingPatch
	if json.Unmarshal([]byte(raw), &p) != nil {
		return
	}
	b.patches.mu.Lock()
	b.patches.pending = p
	b.patches.mu.Unlock()
	if len(p) > 0 {
		b.log.Infof("%d app state change(s) still waiting to be sent", len(p))
	}
}

func sameIndex(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
