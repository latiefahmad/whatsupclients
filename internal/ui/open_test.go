package ui

import (
	"testing"
)

// TestOpenURLTargets checks what openURL accepts and hands on, and that the
// hook TestMain installs keeps every one of them from reaching a browser.
func TestOpenURLTargets(t *testing.T) {
	var got []string
	old := openTargetHook
	openTargetHook = func(target string) bool { got = append(got, target); return true }
	defer func() { openTargetHook = old }()

	if !openURL(donateURL) {
		t.Fatalf("Donate's link was refused: %s", donateURL)
	}
	if len(got) != 1 || got[0] != donateURL {
		t.Errorf("opened %v, want just %q", got, donateURL)
	}

	got = nil
	for _, bad := range []string{"javascript:alert(1)", "file:///etc/passwd", "saweria.co", "https://"} {
		if openURL(bad) {
			t.Errorf("opened %q", bad)
		}
	}
	if len(got) != 0 {
		t.Errorf("invalid links reached the system: %v", got)
	}
}

// TestClickTestOpensNoBrowser checks that the test that clicks all over the
// window leaves its Donate click virtual: without TestMain's hook it would
// start a browser twice on every run.
func TestClickTestOpensNoBrowser(t *testing.T) {
	if openTargetHook == nil {
		t.Fatal("TestMain installs no hook, so tests open browsers")
	}
	var opened []string
	old := openTargetHook
	openTargetHook = func(target string) bool { opened = append(opened, target); return true }
	defer func() { openTargetHook = old }()

	TestClickRedraws(t)

	for _, target := range opened {
		if target == donateURL {
			t.Logf("Donate's click reached openTarget, which the hook stopped: %s", target)
			return
		}
	}
	t.Error("the click test never hit Donate: it no longer covers the title bar")
}
