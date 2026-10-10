package auto

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

var base = time.Date(2026, time.October, 2, 15, 0, 0, 0, time.UTC)

// fake is the demo backend, with messages coming in when a test says.
type fake struct {
	*mock.Backend
	in []model.Event
}

func (f *fake) Poll() []model.Event {
	evs := append(f.Backend.Poll(), f.in...)
	f.in = nil
	return evs
}

type autoTest struct {
	t   *testing.T
	f   *fake
	a   *Backend
	now time.Time
}

func newAutoTest(t *testing.T) *autoTest {
	mock.Clock = func() time.Time { return base }
	at := &autoTest{t: t, f: &fake{Backend: mock.New()}, now: base}
	at.a = Wrap(at.f, func() time.Time { return at.now })
	at.a.Start(func() {})
	at.a.Poll() // online
	return at
}

// sent returns the messages Poll says it sent.
func (at *autoTest) poll() (msgs []*model.Message, notices []string) {
	for _, ev := range at.a.Poll() {
		switch e := ev.(type) {
		case model.MessageEvent:
			if e.Msg.FromMe {
				msgs = append(msgs, e.Msg)
			}
		case model.NoticeEvent:
			notices = append(notices, e.Text)
		}
	}
	return msgs, notices
}

func TestSchedule(t *testing.T) {
	at := newAutoTest(t)
	j := at.a.Schedule(Job{Chat: "rina", At: base.Add(time.Hour), Text: "later", Reply: "rina-2"})
	at.a.Schedule(Job{Chat: "work", At: base.Add(30 * time.Minute), Text: "sooner"})
	if js := at.a.Jobs(""); len(js) != 2 || js[0].Text != "sooner" || len(at.a.Jobs("rina")) != 1 {
		t.Fatalf("jobs %+v", js)
	}
	// They're kept in the prefs.
	if again := Wrap(at.f, func() time.Time { return at.now }); len(again.Jobs("")) != 2 {
		t.Fatal("the jobs weren't saved")
	}
	if ms, _ := at.poll(); len(ms) != 0 {
		t.Fatalf("sent %d early", len(ms))
	}
	at.now = base.Add(32 * time.Minute)
	ms, _ := at.poll()
	if len(ms) != 1 || ms[0].Text != "sooner" || ms[0].ChatID != "work" {
		t.Fatalf("sent %+v", ms)
	}
	// Offline, nothing goes until it's back.
	at.f.in = append(at.f.in, model.ConnEvent{State: model.StateConnecting})
	at.now = base.Add(time.Hour + 3*time.Minute) // a little late is fine
	if ms, _ := at.poll(); len(ms) != 0 {
		t.Fatal("sent while offline")
	}
	at.f.in = append(at.f.in, model.ConnEvent{State: model.StateOnline})
	ms, _ = at.poll()
	if len(ms) != 1 || ms[0].Text != "later" || ms[0].Quote == nil {
		t.Fatalf("sent %+v", ms)
	}
	if at.a.Cancel(j.ID) || len(at.a.Jobs("")) != 0 || at.f.Pref(prefJobs) != "" {
		t.Fatal("the queue isn't empty")
	}
}

func TestScheduleTooLate(t *testing.T) {
	at := newAutoTest(t)
	at.a.Schedule(Job{Chat: "rina", At: base.Add(time.Hour), Text: "good morning everyone, coffee's on me today", Shown: ""})
	at.a.Schedule(Job{Chat: "rina", At: base.Add(3 * time.Hour), Text: "later"})
	// The app was closed: it opens two hours late.
	at.now = base.Add(3 * time.Hour)
	ms, ns := at.poll()
	if len(ms) != 1 || ms[0].Text != "later" {
		t.Fatalf("sent %+v", ms)
	}
	want := "Not sent, the app was closed or offline at 16:00: \"good morning everyone, coffee's on me t…\""
	if len(ns) != 1 || ns[0] != want {
		t.Fatalf("notices %q, want %q", ns, want)
	}
	if len(at.a.Jobs("")) != 0 {
		t.Fatal("the late message stayed queued")
	}
}

func TestCancelAndSendNow(t *testing.T) {
	at := newAutoTest(t)
	a := at.a.Schedule(Job{Chat: "rina", At: base.Add(time.Hour), Text: "a"})
	b := at.a.Schedule(Job{Chat: "rina", At: base.Add(time.Hour), Text: "b"})
	if !at.a.Cancel(a.ID) {
		t.Fatal("couldn't cancel")
	}
	if m := at.a.SendNow(b.ID); m == nil || m.Text != "b" {
		t.Fatalf("sent %v", m)
	}
	if at.a.SendNow(b.ID) != nil || len(at.a.Jobs("")) != 0 {
		t.Fatal("sent twice")
	}
}

// mentioned is a mention as backends resolve it, marked as notifying you.
func mentioned(name string) string {
	return "⁨" + string(model.MentionNotifies) + "@" + name + "⁩"
}

func replyToMe(id, chat, sender string, t time.Time) model.Event {
	ev := incoming(id, chat, sender, "agreed", t).(model.MessageEvent)
	ev.Msg.Quote = &model.Quote{Sender: "You", Text: "ship it?"}
	return ev
}

func incoming(id, chat, sender, text string, t time.Time) model.Event {
	return model.MessageEvent{New: true, Msg: &model.Message{ID: id, ChatID: chat, Sender: sender, SenderID: sender,
		Text: text, Time: t}}
}

func TestAway(t *testing.T) {
	at := newAutoTest(t)
	at.a.SetAway("lunch")
	at.now = base.Add(10 * time.Minute)
	at.f.in = append(at.f.in,
		incoming("x1", "rina", "rina@lid", "hey", at.now),
		incoming("x2", "rina", "rina@lid", "you there?", at.now),    // told already
		incoming("x3", "work", "budi@lid", "lunch anyone?", at.now), // a group, not to you
		incoming("x4", "work", "dewi@lid", "ping "+mentioned("You"), at.now),
		incoming("x6", "work", "andre@lid", mentioned("all")+" standup", at.now), // not @all
		replyToMe("x7", "work", "clara@lid", at.now),                             // nor a reply in a group
		incoming("x5", "rina", "rina@lid", "old", base.Add(-time.Hour)),          // from before
	)
	ms, _ := at.poll()
	if len(ms) != 2 || ms[0].ChatID != "rina" || ms[1].ChatID != "work" || ms[1].Quote == nil || ms[1].Quote.ID != "x4" {
		t.Fatalf("replied %+v", ms)
	}
	if want := "💤 AFK: lunch"; ms[0].Text != want || ms[0].Quote == nil {
		t.Fatalf("reply %q, want %q", ms[0].Text, want)
	}
	// Its own replies coming back don't end it.
	at.f.in = append(at.f.in, model.MessageEvent{Msg: ms[0]})
	if _, ns := at.poll(); len(ns) != 0 || at.a.Away() == nil {
		t.Fatal("AFK ended by its own reply")
	}
	// It's kept in the prefs.
	if again := Wrap(at.f, func() time.Time { return at.now }); again.Away() == nil || len(again.Away().Told) != 2 {
		t.Fatal("AFK wasn't saved")
	}
	// Sending a message ends it.
	at.a.Send("rina", model.Draft{Text: "back"})
	if _, ns := at.poll(); len(ns) != 1 || !strings.Contains(ns[0], "2 people") || at.a.Away() != nil {
		t.Fatalf("notices %q", ns)
	}
	if at.f.Pref(prefAway) != "" {
		t.Fatal("AFK stayed in the prefs")
	}
}

func TestAFKAllow(t *testing.T) {
	at := newAutoTest(t)
	at.a.SetAway("lunch")
	if l := at.a.AllowList(); l.Only || len(l.Members) != 0 {
		t.Fatalf("allowlist starts %+v", l)
	}
	// Rina by her picked ID, Agus by his typed number.
	if !at.a.AllowMember("rina@lid", "Rina") || !at.a.AllowMember("6281355501942", "+62 813-5550-1942") {
		t.Fatal("adds didn't stick")
	}
	if at.a.AllowMember("rina@lid", "Rina") {
		t.Fatal("dup added twice")
	}
	at.a.SetAllowOnly(true)
	at.now = base.Add(10 * time.Minute)
	at.f.in = append(at.f.in,
		incoming("a1", "rina", "rina@lid", "hey", at.now),
		incoming("a2", "rina", "rina@lid", "again", at.now),                  // told already
		incoming("a3", "agus", "agus", "morning", at.now),                    // typed number, via contacts
		incoming("a4", "fitri", "fitri", "hi", at.now),                       // not listed
		incoming("a5", "work", "budi@lid", "ping "+mentioned("You"), at.now), // group, not listed
	)
	ms, _ := at.poll()
	if len(ms) != 2 || ms[0].ChatID != "rina" || ms[1].ChatID != "agus" {
		t.Fatalf("replied %+v", ms)
	}
	// Unallowing by digits drops Agus; names work too. The list and the
	// mode persist.
	if n := at.a.UnallowMember("6281355501942"); n != 1 {
		t.Fatalf("unallowed %d", n)
	}
	if !at.a.AllowMember("yoga@lid", "Yoga") || at.a.UnallowMember("yoga") != 1 {
		t.Fatal("unallow by name didn't drop Yoga")
	}
	if again := Wrap(at.f, func() time.Time { return at.now }); !again.AllowList().Only ||
		len(again.AllowList().Members) != 1 || again.AllowList().Members[0].Name != "Rina" {
		t.Fatalf("allowlist saved %+v", again.AllowList())
	}
	// Off again, everyone gets it.
	at.a.SetAllowOnly(false)
	at.f.in = append(at.f.in, incoming("a6", "fitri", "fitri", "hi again", at.now))
	if ms, _ := at.poll(); len(ms) != 1 || ms[0].ChatID != "fitri" {
		t.Fatalf("replied %+v", ms)
	}
}

func TestResolvePerson(t *testing.T) {
	at := newAutoTest(t)
	members := []model.Member{
		{ID: "budi@lid", Name: "Budi Santoso"},
		{ID: "bima@lid", Name: "Bima Sakti"},
		{ID: "me@lid", Name: "Me", Me: true},
	}
	for _, c := range []struct {
		text, id, name, err string
	}{
		{"@Budi Santoso", "budi@lid", "Budi Santoso", ""},
		{"bima", "bima@lid", "Bima Sakti", ""},
		{"Bu", "budi@lid", "Budi Santoso", ""},
		{"B", "", "", "could be Budi Santoso or Bima Sakti"},
		{"Me", "", "", "no contact named"},
		{"Agus Wibowo", "agus", "Agus Wibowo", ""},
		{"Fit", "fitri", "Fitri Handayani", ""},
		{"+62 813-5550-1942", "6281355501942", "Agus Wibowo", ""},
		{"+62 800 0000", "628000000", "+62 800 0000", ""},
		{"Nobody Here", "", "", "no contact named"},
		{"", "", "", "name someone"},
	} {
		id, name, err := ResolvePerson(at.f, members, c.text)
		if id != c.id || name != c.name || (err == nil) != (c.err == "") ||
			(err != nil && !strings.Contains(err.Error(), c.err)) {
			t.Errorf("ResolvePerson(%q) = %q, %q, %v", c.text, id, name, err)
		}
	}
}

func TestAFKHours(t *testing.T) {
	at := newAutoTest(t)
	if h := at.a.Hours(); h.On {
		t.Fatalf("hours start %+v", h)
	}
	if _, err := ParseHour("21:00"); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"21", "9", "21:00", "07:00"} {
		if _, err := ParseHour(s); err != nil {
			t.Errorf("ParseHour(%q): %v", s, err)
		}
	}
	for _, s := range []string{"", "21:00:00", "24:00", "21:60", "noon", "21:00-07:00"} {
		if _, err := ParseHour(s); err == nil {
			t.Errorf("ParseHour(%q) passed", s)
		}
	}
	if FormatHour(21*60+7) != "21:07" || FormatHour(9*60) != "9:00" {
		t.Errorf("FormatHour %q %q", FormatHour(21*60+7), FormatHour(9*60))
	}
	// Overnight: replies go 21:00-07:00.
	at.a.SetHours(true, 21*60, 7*60)
	at.now = time.Date(2026, time.October, 2, 0, 1, 0, 0, time.UTC)
	at.a.SetAway("night")
	for tm, want := range map[string]bool{
		"20:59": false, "21:00": true, "23:30": true, "06:59": true, "07:00": false, "12:00": false,
	} {
		h, m, _ := strings.Cut(tm, ":")
		at.now = time.Date(2026, time.October, 2, atoi(h), atoi(m), 0, 0, time.UTC)
		at.f.in = append(at.f.in, incoming("h-"+tm, "c"+tm, "s"+tm, "hey "+tm, at.now))
		ms, _ := at.poll()
		if (len(ms) == 1) != want {
			t.Errorf("%s: replied %v, want %v", tm, len(ms) == 1, want)
		}
	}
	// Day span: 09:00-17:00.
	at.a.SetHours(true, 9*60, 17*60)
	for tm, want := range map[string]bool{"08:59": false, "09:00": true, "12:00": true, "17:00": false} {
		h, m, _ := strings.Cut(tm, ":")
		at.now = time.Date(2026, time.October, 2, atoi(h), atoi(m), 0, 0, time.UTC)
		at.f.in = append(at.f.in, incoming("d-"+tm, "d"+tm, "t"+tm, "hey "+tm, at.now))
		ms, _ := at.poll()
		if (len(ms) == 1) != want {
			t.Errorf("%s: replied %v, want %v", tm, len(ms) == 1, want)
		}
	}
	// Off again, and kept in the prefs.
	at.a.SetHours(false, 0, 0)
	at.now = base.Add(10 * time.Minute)
	at.f.in = append(at.f.in, incoming("h-off", "rina", "rina@lid", "hey", at.now))
	if ms, _ := at.poll(); len(ms) != 1 {
		t.Fatalf("replied %+v", ms)
	}
	at.a.SetHours(true, 21*60, 7*60)
	if again := Wrap(at.f, func() time.Time { return at.now }); again.Hours() != (AFKHours{On: true, From: 21 * 60, To: 7 * 60}) {
		t.Fatalf("hours saved %+v", again.Hours())
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func TestAwayReasonKept(t *testing.T) {
	at := newAutoTest(t)
	at.a.SetAway("lunch")
	since := at.a.Away().Since
	at.a.SetAwayReason("back at 2")
	if w := at.a.Away(); w.Reason != "back at 2" || !w.Since.Equal(since) {
		t.Fatalf("away %+v", w)
	}
	if again := Wrap(at.f, func() time.Time { return at.now }); again.Away() == nil ||
		again.Away().Reason != "back at 2" {
		t.Fatal("reason wasn't saved")
	}
}

func TestAwayEndsFromAnotherDevice(t *testing.T) {
	at := newAutoTest(t)
	at.a.SetAway("")
	at.f.in = append(at.f.in, model.MessageEvent{Msg: &model.Message{ID: "p1", ChatID: "rina", FromMe: true,
		Time: base.Add(time.Minute)}})
	if _, ns := at.poll(); len(ns) != 1 || ns[0] != "Welcome back! You're no longer AFK." {
		t.Fatalf("notices %q", ns)
	}
}

func TestAwayText(t *testing.T) {
	if got := AwayText(&Away{Reason: "lunch"}); got != "💤 AFK: lunch" {
		t.Errorf("AwayText = %q", got)
	}
	if got := AwayText(&Away{}); got != "💤 AFK" {
		t.Errorf("AwayText without reason = %q", got)
	}
	if got := AwayPlain(&Away{Reason: "lunch"}); got != "💤 AFK: lunch" {
		t.Errorf("AwayPlain = %q", got)
	}
}
