package wa

import (
	"fmt"
	"strings"
	"testing"

	"github.com/polymorfa/hypermeow/appstate"
	waLog "github.com/polymorfa/hypermeow/util/log"
)

// TestShortHashDistinguishesStates checks that a log line can tell two app
// state hashes apart without a page of hex.
func TestShortHashDistinguishesStates(t *testing.T) {
	var a, b [128]byte
	a[0], a[127] = 0xab, 0xcd
	b[0], b[127] = 0xab, 0xce
	if shortHash(a) == shortHash(b) {
		t.Errorf("hashes differ but both print as %s", shortHash(a))
	}
	if shortHash(a) != "ab000000..000000cd" {
		t.Errorf("unexpected form: %s", shortHash(a))
	}
}

// captureLogger keeps every formatted line, for a test to read back.
type captureLogger struct{ lines []string }

func (l *captureLogger) add(f string, a ...any) {
	l.lines = append(l.lines, fmt.Sprintf(f, a...))
}
func (l *captureLogger) Warnf(f string, a ...any)  { l.add(f, a...) }
func (l *captureLogger) Errorf(f string, a ...any) { l.add(f, a...) }
func (l *captureLogger) Infof(f string, a ...any)  { l.add(f, a...) }
func (l *captureLogger) Debugf(f string, a ...any) { l.add(f, a...) }
func (l *captureLogger) Sub(string) waLog.Logger   { return l }

// TestDiagSilentWithoutDebug checks that the diagnostics log nothing unless
// the app runs with -debug.
func TestDiagSilentWithoutDebug(t *testing.T) {
	log := &captureLogger{}
	b := newQueueTestBackend(t)
	b.log = log
	b.debugLog = false

	b.noteAppStateSend(appstate.WAPatchRegularHigh, 1)
	b.noteAppStateResult(appstate.WAPatchRegularHigh, "v1", nil)
	b.noteAppStateSyncError(appstate.WAPatchRegularHigh, errTest)
	b.noteRecoveryRefused(appstate.WAPatchRegularHigh)

	if len(log.lines) != 0 {
		t.Errorf("logged without -debug: %q", log.lines)
	}
}

// TestDiagSpeaksWithDebug checks that -debug does record them, or the switch
// would hide the very thing it was added for.
func TestDiagSpeaksWithDebug(t *testing.T) {
	log := &captureLogger{}
	b := newQueueTestBackend(t)
	b.log = log
	b.debugLog = true

	b.noteAppStateSyncError(appstate.WAPatchRegularHigh, errTest)

	if len(log.lines) == 0 {
		t.Fatal("nothing logged with -debug")
	}
	if !strings.Contains(log.lines[0], "app state diag") {
		t.Errorf("not a diagnostic line: %q", log.lines[0])
	}
}

type simpleErr struct{}

func (simpleErr) Error() string { return "mismatching LTHash" }

var errTest = simpleErr{}
