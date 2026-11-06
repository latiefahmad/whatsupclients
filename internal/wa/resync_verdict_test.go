package wa

import (
	"testing"

	"github.com/polymorfa/hypermeow/appstate"
)

// TestResyncVerdict covers the four cases a collection can be in when a
// connect starts. The one that broke here was "never marked, no state":
// kept out of the pass and so never fetched, which left regular_high at
// version 0 and every change to it queued and never sent.
func TestResyncVerdict(t *testing.T) {
	cases := []struct {
		name             string
		marked, hasState bool
		want             resyncVerdict
	}{
		{"finished last time", true, true, resyncSealed},
		{"marked but stored nothing", true, false, resyncDropMark},
		{"never reached, has state", false, true, resyncReopen},
		{"never reached, no state", false, false, resyncReopen},
	}
	for _, c := range cases {
		if got := resyncVerdictFor(c.marked, c.hasState); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// TestNoStateReopensThePass pins the case that actually happened: with the
// overall marker set, a collection that was never marked and has no state
// must still reopen the pass, or nothing fetches it ever again.
func TestNoStateReopensThePass(t *testing.T) {
	if got := resyncVerdictFor(false, false); got == resyncSealed {
		t.Fatal("a collection with no state was left sealed; it can never sync")
	}
	// And the collection that broke, by name, is one the pass walks.
	found := false
	for _, n := range appstate.AllPatchNames {
		if n == appstate.WAPatchRegularHigh {
			found = true
		}
	}
	if !found {
		t.Fatal("regular_high is not in AllPatchNames; the pass never sees it")
	}
}
