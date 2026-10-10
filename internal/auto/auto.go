// Package auto sends messages for you: the ones you schedule (/schedule)
// and, while you're away (/afk), a reply to whoever messages or mentions
// you. It wraps the Backend, so it sees every event and every message you
// send, and keeps its state in the backend's prefs: it goes on with the
// window closed, and after a restart.
//
// In ghost mode (model.PrefGhost) it refuses everything you'd send
// yourself, with a notice: messages, reactions, votes, edits, deleting or
// pinning for everyone and status updates. What it sends for you still
// goes, since it sends through the backend it wraps.
//
// Like the Backend, it's used from the UI goroutine only; its timer just
// pokes the backend's notify function, and the work happens in Poll.
package auto

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

// Pref keys for the state it keeps.
const (
	prefJobs  = "auto_scheduled"
	prefAway  = "auto_away"
	prefAllow = "auto_afk_allow"
	prefHours = "auto_afk_hours"
)

// Job is a scheduled message.
type Job struct {
	ID   string    `json:"id"`
	Chat string    `json:"chat"`
	At   time.Time `json:"at"`
	// Text is the message as sent, with mentions as "@<user>", and Shown
	// as you typed it ("@Budi"), for listing it.
	Text          string   `json:"text"`
	Shown         string   `json:"shown,omitempty"`
	Mentions      []string `json:"mentions,omitempty"`
	MentionAll    bool     `json:"all,omitempty"`
	MentionAdmins bool     `json:"admins,omitempty"`
	// Reply is the ID of the message it replies to, or "".
	Reply string `json:"reply,omitempty"`
}

// Away is the AFK state.
type Away struct {
	Reason string    `json:"reason,omitempty"`
	Since  time.Time `json:"since"`
	// Told lists who got the reply, as "chat|sender": each once.
	Told []string `json:"told,omitempty"`
}

// AFKMember is one contact the AFK reply may go to.
type AFKMember struct {
	// ID is the JID as picked, or the digits for a typed number.
	ID string `json:"id"`
	// Name is as picked or typed, for listing.
	Name string `json:"name,omitempty"`
}

// AFKAllow is who the AFK reply goes to (Settings > AFK list). Without Only, or
// with no members, everyone answers() names gets it.
type AFKAllow struct {
	Only    bool        `json:"only,omitempty"`
	Members []AFKMember `json:"members,omitempty"`
}

// AFKHours confines AFK replies to a daily span, in minutes since
// midnight. Overnight spans (from after to) wrap past midnight.
type AFKHours struct {
	On   bool `json:"on,omitempty"`
	From int  `json:"from,omitempty"`
	To   int  `json:"to,omitempty"`
}

// ParseHour reads "21:00" or "21" as minutes since midnight.
func ParseHour(s string) (int, error) {
	if h, m, ok := strings.Cut(s, ":"); ok {
		hh, err1 := strconv.Atoi(h)
		mm, err2 := strconv.Atoi(m)
		if err1 != nil || err2 != nil || hh < 0 || hh > 23 || mm < 0 || mm > 59 {
			return 0, errors.New("type an hour like 21:00")
		}
		return hh*60 + mm, nil
	}
	if hh, err := strconv.Atoi(s); err == nil && hh >= 0 && hh <= 23 {
		return hh * 60, nil
	}
	return 0, errors.New("type an hour like 21:00")
}

// FormatHour shows minutes since midnight as "21:00".
func FormatHour(m int) string {
	out := strconv.Itoa(m/60) + ":"
	if m%60 < 10 {
		out += "0"
	}
	return out + strconv.Itoa(m%60)
}

// Backend is a model.Backend that also sends scheduled messages and AFK
// replies.
type Backend struct {
	model.Backend
	now    func() time.Time
	jobs   []Job // by At
	away   *Away
	allow  AFKAllow
	hours  AFKHours
	meID   string
	online bool
	// ours are the IDs of the messages it sent: they don't end AFK.
	ours map[string]bool
	// groups caches which chats are groups (see isGroup).
	groups map[string]bool
	// byPhone maps contact digits to chat IDs, for matching typed
	// numbers against LID-keyed chats; nil until first use.
	byPhone map[string]string
	out     []model.Event // for the next Poll
	notify  func()
	timer   *time.Timer
	seq     int
	ver     int // bumped as the jobs change
	ghost   bool
}

// Wrap returns b with scheduled messages and AFK, reading the time from
// now.
func Wrap(b model.Backend, now func() time.Time) *Backend {
	a := &Backend{Backend: b, now: now, ours: map[string]bool{}, groups: map[string]bool{},
		ghost: b.Pref(model.PrefGhost) == "on"}
	if s := b.Pref(prefJobs); s != "" {
		json.Unmarshal([]byte(s), &a.jobs)
	}
	if s := b.Pref(prefAway); s != "" {
		var w Away
		if json.Unmarshal([]byte(s), &w) == nil {
			a.away = &w
		}
	}
	if s := b.Pref(prefAllow); s != "" {
		json.Unmarshal([]byte(s), &a.allow)
	}
	if s := b.Pref(prefHours); s != "" {
		json.Unmarshal([]byte(s), &a.hours)
	}
	return a
}

// Start starts the backend.
func (a *Backend) Start(notify func()) {
	a.notify = notify
	a.Backend.Start(notify)
	a.arm()
}

// Close stops the backend.
func (a *Backend) Close() {
	if a.timer != nil {
		a.timer.Stop()
	}
	a.Backend.Close()
}

// Poll returns the backend's events, plus the messages it sent and its
// notices.
func (a *Backend) Poll() []model.Event {
	evs := a.Backend.Poll()
	for _, ev := range evs {
		switch e := ev.(type) {
		case model.ConnEvent:
			if e.MeID != "" {
				a.meID = e.MeID
			}
			a.online = e.State == model.StateOnline
		case model.ChatsEvent:
			for _, c := range e.Chats {
				a.groups[c.ID] = c.IsGroup
			}
		case model.ChatEvent:
			a.groups[e.Chat.ID] = e.Chat.IsGroup
		case model.MessageEvent:
			a.seen(e)
		}
	}
	a.sendDue()
	if len(a.out) > 0 {
		evs = append(evs, a.out...)
		a.out = nil
	}
	return evs
}

// Schedule adds j to the queue, giving it an ID, and returns it.
func (a *Backend) Schedule(j Job) Job {
	a.seq++
	j.ID = strconv.FormatInt(a.now().UnixNano(), 36) + "-" + strconv.Itoa(a.seq)
	i := 0
	for i < len(a.jobs) && !a.jobs[i].At.After(j.At) {
		i++
	}
	a.jobs = append(a.jobs[:i:i], append([]Job{j}, a.jobs[i:]...)...)
	a.saveJobs()
	a.arm()
	return j
}

// Jobs lists the messages scheduled in chat ("" for every chat), soonest
// first.
func (a *Backend) Jobs(chat string) []Job {
	var out []Job
	for _, j := range a.jobs {
		if chat == "" || j.Chat == chat {
			out = append(out, j)
		}
	}
	return out
}

// Cancel takes the job with id off the queue, and reports whether it was
// still on it.
func (a *Backend) Cancel(id string) bool {
	for i, j := range a.jobs {
		if j.ID == id {
			a.jobs = append(a.jobs[:i:i], a.jobs[i+1:]...)
			a.saveJobs()
			a.arm()
			return true
		}
	}
	return false
}

// SendNow sends the job with id right away, and returns the message (nil
// when the job is gone).
func (a *Backend) SendNow(id string) *model.Message {
	for i, j := range a.jobs {
		if j.ID == id {
			a.jobs = append(a.jobs[:i:i], a.jobs[i+1:]...)
			a.saveJobs()
			a.arm()
			return a.send(j)
		}
	}
	return nil
}

// lateBy is how late a scheduled message may still go: past it (the app
// was closed or offline), it's dropped with a notice instead.
const lateBy = 5 * time.Minute

// sendDue sends the jobs whose time has come, once online, and drops the
// ones too late to send.
func (a *Backend) sendDue() {
	if !a.online {
		return
	}
	now := a.now()
	n := 0
	for n < len(a.jobs) && !a.jobs[n].At.After(now) {
		n++
	}
	if n == 0 {
		return
	}
	due := a.jobs[:n:n]
	a.jobs = append([]Job(nil), a.jobs[n:]...)
	a.saveJobs()
	for _, j := range due {
		if now.Sub(j.At) > lateBy {
			a.out = append(a.out, model.NoticeEvent{Text: MissedText(j, now)})
			continue
		}
		if m := a.send(j); m != nil {
			a.out = append(a.out, model.MessageEvent{Msg: m})
		}
	}
	a.arm()
}

func (a *Backend) send(j Job) *model.Message {
	d := model.Draft{Text: j.Text, Mentions: j.Mentions, MentionAll: j.MentionAll, MentionAdmins: j.MentionAdmins}
	if j.Reply != "" {
		if ms := a.Backend.MessagesFrom(j.Chat, j.Reply, 1); len(ms) > 0 && ms[0].ID == j.Reply {
			d.Reply = ms[0]
		}
	}
	m := a.Backend.Send(j.Chat, d)
	if m != nil {
		a.ours[m.ID] = true
	}
	return m
}

// MissedText says that j wasn't sent because it was too late.
func MissedText(j Job, now time.Time) string {
	text := []rune(j.Shown)
	if len(text) == 0 {
		text = []rune(j.Text)
	}
	if len(text) > 40 {
		text = append(text[:39], '…')
	}
	at := j.At.In(now.Location())
	when := at.Format("15:04")
	if y, m, d := at.Date(); y != now.Year() || m != now.Month() || d != now.Day() {
		when = at.Format("Mon 2 Jan 15:04")
	}
	return "Not sent, the app was closed or offline at " + when + ": \"" + string(text) + "\""
}

// arm sets the timer for the next job. It wakes at least hourly, in case
// the clock jumped (the computer slept).
func (a *Backend) arm() {
	if a.notify == nil {
		return
	}
	if a.timer != nil {
		a.timer.Stop()
		a.timer = nil
	}
	if len(a.jobs) == 0 {
		return
	}
	d := min(max(a.jobs[0].At.Sub(a.now()), 0), time.Hour)
	a.timer = time.AfterFunc(d, a.notify)
}

// Version changes whenever the scheduled messages do.
func (a *Backend) Version() int { return a.ver }

func (a *Backend) saveJobs() {
	a.ver++
	if len(a.jobs) == 0 {
		a.Backend.SetPref(prefJobs, "")
		return
	}
	if b, err := json.Marshal(a.jobs); err == nil {
		a.Backend.SetPref(prefJobs, string(b))
	}
}

// Away returns the AFK state, or nil when you're not away.
func (a *Backend) Away() *Away { return a.away }

// SetAway marks you away for reason, from now on, until you send a
// message.
func (a *Backend) SetAway(reason string) {
	a.away = &Away{Reason: reason, Since: a.now()}
	a.saveAway()
}

// SetAwayReason changes the away message, keeping when it started.
func (a *Backend) SetAwayReason(reason string) {
	if a.away == nil {
		return
	}
	a.away.Reason = reason
	a.saveAway()
}

// Back ends AFK, and returns how many got the reply.
func (a *Backend) Back() int {
	if a.away == nil {
		return 0
	}
	n := len(a.away.Told)
	a.away = nil
	a.saveAway()
	return n
}

func (a *Backend) saveAway() {
	if a.away == nil {
		a.Backend.SetPref(prefAway, "")
		return
	}
	if b, err := json.Marshal(a.away); err == nil {
		a.Backend.SetPref(prefAway, string(b))
	}
}

// Digits keeps the decimal digits of s: what phone numbers and JIDs
// compare by when their form (LID or phone) differs.
func Digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ResolvePerson finds who text calls for the AFK list: a group member
// by exact then unique-prefix name (never you), else a saved contact
// the same way, else a typed phone number. Numbers take the contact's
// name when known.
func ResolvePerson(b model.Backend, members []model.Member, text string) (id, name string, err error) {
	t := strings.TrimSpace(strings.TrimPrefix(text, "@"))
	if t == "" {
		return "", "", errors.New("name someone: @mention, call or type them")
	}
	var hits []model.Member
	for _, m := range members {
		if m.Me || m.ID == "" {
			continue
		}
		if strings.EqualFold(m.Name, t) {
			return m.ID, m.Name, nil
		}
		if strings.HasPrefix(strings.ToLower(m.Name), strings.ToLower(t)) {
			hits = append(hits, m)
		}
	}
	if len(hits) == 1 {
		return hits[0].ID, hits[0].Name, nil
	}
	if len(hits) > 1 {
		ns := make([]string, len(hits))
		for i, m := range hits {
			ns[i] = m.Name
		}
		return "", "", errors.New(t + " could be " + strings.Join(ns, " or ") + "; be more specific")
	}
	if id, name, n := matchContact(b.Contacts(), t); n == 1 {
		return id, name, nil
	} else if n > 1 {
		return "", "", errors.New(t + " could be several contacts; be more specific")
	}
	if d, ok := phoneDigits(t); ok {
		name := t
		for _, ct := range b.Contacts() {
			if Digits(ct.Phone) == d {
				name = ct.Name
				break
			}
		}
		return d, name, nil
	}
	return "", "", errors.New("no contact named " + t + "; type their number with its country code")
}

// matchContact finds saved contact t calls, by exact then unique-prefix
// name. It reports how many prefix names fit.
func matchContact(contacts []*model.Contact, t string) (id, name string, n int) {
	var hit *model.Contact
	for _, ct := range contacts {
		if strings.EqualFold(ct.Name, t) {
			return ct.ID, ct.Name, 1
		}
		if strings.HasPrefix(strings.ToLower(ct.Name), strings.ToLower(t)) {
			hit, n = ct, n+1
		}
	}
	if n == 1 {
		return hit.ID, hit.Name, 1
	}
	return "", "", n
}

// phoneDigits is t's digits when it reads as a phone number.
func phoneDigits(t string) (string, bool) {
	for _, r := range strings.TrimRight(t, ",") {
		if !strings.ContainsRune("+0123456789-(). ", r) {
			return "", false
		}
	}
	if d := Digits(t); len(d) >= 7 {
		return d, true
	}
	return "", false
}

// AllowList is who the AFK reply may go to.
func (a *Backend) AllowList() AFKAllow {
	out := a.allow
	out.Members = append([]AFKMember(nil), a.allow.Members...)
	return out
}

// SetAllowOnly confines the AFK reply to the listed contacts when on.
func (a *Backend) SetAllowOnly(on bool) {
	a.allow.Only = on
	a.saveAllow()
}

// AllowMember adds id (a JID, or digits for a typed number) to the AFK
// list under name, and reports whether it wasn't there.
func (a *Backend) AllowMember(id, name string) bool {
	if id == "" {
		return false
	}
	if !strings.Contains(id, "@") {
		// A typed number, kept as digits; anything else stays as is.
		if d := Digits(id); len(d) >= 7 {
			id = d
		}
	}
	// The same human in another ID form (a picked contact vs their
	// typed number) is already there.
	key := a.AllowKey(id)
	for _, m := range a.allow.Members {
		if a.AllowKey(m.ID) == key {
			return false
		}
	}
	if name == "" {
		name = id
	}
	a.allow.Members = append(a.allow.Members, AFKMember{ID: id, Name: name})
	a.byPhone = nil
	a.saveAllow()
	return true
}

// AllowKey is the canonical identity of an allowlist ID: the chat ID
// of the contact with its number, or the ID itself.
func (a *Backend) AllowKey(id string) string {
	if d := Digits(id); len(d) >= 7 {
		if cid, ok := a.contactID(d); ok {
			return cid
		}
	}
	return id
}

// UnallowMember drops the listed contacts matching match (digits, JID
// or name), and reports how many went.
func (a *Backend) UnallowMember(match string) int {
	d := Digits(match)
	kept := a.allow.Members[:0]
	for _, m := range a.allow.Members {
		md := Digits(m.ID)
		if m.ID == match || (len(d) >= 7 && len(md) >= 7 && md == d) ||
			strings.EqualFold(m.Name, match) {
			continue
		}
		kept = append(kept, m)
	}
	n := len(a.allow.Members) - len(kept)
	if n > 0 {
		a.allow.Members = append([]AFKMember(nil), kept...)
		a.byPhone = nil
		a.saveAllow()
	}
	return n
}

func (a *Backend) saveAllow() {
	if !a.allow.Only && len(a.allow.Members) == 0 {
		a.Backend.SetPref(prefAllow, "")
		return
	}
	if b, err := json.Marshal(a.allow); err == nil {
		a.Backend.SetPref(prefAllow, string(b))
	}
}

// Hours is when AFK replies go: all day unless confined.
func (a *Backend) Hours() AFKHours { return a.hours }

// SetHours confines AFK replies to from..to minutes since midnight
// (overnight when to is earlier), or lifts the span when off.
func (a *Backend) SetHours(on bool, from, to int) {
	a.hours = AFKHours{On: on, From: from, To: to}
	if !on {
		a.Backend.SetPref(prefHours, "")
		return
	}
	if b, err := json.Marshal(a.hours); err == nil {
		a.Backend.SetPref(prefHours, string(b))
	}
}

// inWindow reports whether replies go at now: always, unless confined
// to hours outside it.
func (a *Backend) inWindow(now time.Time) bool {
	h := a.hours
	if !h.On || h.From == h.To {
		return true
	}
	m := now.Hour()*60 + now.Minute()
	if h.From < h.To {
		return m >= h.From && m < h.To
	}
	return m >= h.From || m < h.To
}

// allowed reports whether m's sender may get the AFK reply: anyone,
// unless the reply is confined to the listed contacts.
func (a *Backend) allowed(m *model.Message) bool {
	if !a.allow.Only {
		return true
	}
	if len(a.allow.Members) == 0 {
		return false
	}
	for _, id := range []string{m.ChatID, m.SenderID, m.Sender} {
		if a.allowMatch(id) {
			return true
		}
	}
	// Contacts may have arrived since the map was built.
	a.byPhone = nil
	for _, id := range []string{m.ChatID, m.SenderID, m.Sender} {
		if a.allowMatch(id) {
			return true
		}
	}
	return false
}

// allowMatch reports whether id is a listed contact: the ID itself, its
// digits, or the chat ID of the contact with its number.
func (a *Backend) allowMatch(id string) bool {
	if id == "" {
		return false
	}
	d := Digits(id)
	for _, m := range a.allow.Members {
		md := Digits(m.ID)
		if m.ID == id || (len(d) >= 7 && len(md) >= 7 && md == d) {
			return true
		}
		// The member's number may map to another ID form (a typed
		// number to a LID-keyed chat): compare through the contacts.
		if len(md) >= 7 {
			if cid, ok := a.contactID(md); ok && (cid == id || (len(d) >= 7 && Digits(cid) == d)) {
				return true
			}
		}
	}
	return false
}

// contactID maps contact digits to the contact's chat ID, building the
// map on first use.
func (a *Backend) contactID(digits string) (string, bool) {
	if digits == "" {
		return "", false
	}
	if a.byPhone == nil {
		a.byPhone = map[string]string{}
		for _, c := range a.Backend.Contacts() {
			if d := Digits(c.Phone); len(d) >= 7 {
				a.byPhone[d] = c.ID
			}
			if d := Digits(c.ID); len(d) >= 7 {
				if _, ok := a.byPhone[d]; !ok {
					a.byPhone[d] = c.ID
				}
			}
		}
	}
	id, ok := a.byPhone[digits]
	return id, ok
}

// seen looks at a message coming in or going out.
func (a *Backend) seen(e model.MessageEvent) {
	w, m := a.away, e.Msg
	if w == nil || m == nil || m.Time.Before(w.Since) {
		return
	}
	if m.FromMe {
		if !a.ours[m.ID] {
			a.cameBack() // sent from another device
		}
		return
	}
	if !e.New || !a.answers(m) || !a.inWindow(a.now()) {
		return
	}
	who := m.SenderID
	if who == "" {
		who = m.Sender
	}
	key := m.ChatID + "|" + who
	for _, k := range w.Told {
		if k == key {
			return
		}
	}
	w.Told = append(w.Told, key)
	a.saveAway()
	if r := a.Backend.Send(m.ChatID, model.Draft{Text: AwayText(w), Reply: m}); r != nil {
		a.ours[r.ID] = true
		a.out = append(a.out, model.MessageEvent{Msg: r})
	}
}

// answers reports whether m gets the AFK reply: a message to you, or one
// in a group that mentions you by name (not @all, nor a reply to you).
// Channels, status and broadcasts don't. When the reply is confined to
// the listed contacts (confined on the AFK list page), only they do.
func (a *Backend) answers(m *model.Message) bool {
	id := m.ChatID
	switch {
	case strings.HasSuffix(id, "@newsletter"), strings.HasSuffix(id, "@broadcast"):
		return false
	case a.isGroup(id):
		return mentionsMe(m.Text) && a.allowed(m)
	}
	return a.allowed(m)
}

// mentionsMe reports whether text, as a backend resolved it, mentions you:
// a mention marked model.MentionNotifies that isn't "@all".
func mentionsMe(text string) bool {
	mark := string(model.MentionNotifies)
	for {
		i := strings.Index(text, mark)
		if i < 0 {
			return false
		}
		text = text[i+len(mark):]
		if !strings.HasPrefix(text, "@all\u2069") {
			return true
		}
	}
}

// isGroup reports whether chat id is a group, asking the backend about
// chats it hasn't heard of.
func (a *Backend) isGroup(id string) bool {
	if strings.HasSuffix(id, "@g.us") {
		return true
	}
	g, ok := a.groups[id]
	if !ok {
		for _, c := range a.Backend.Chats() {
			a.groups[c.ID] = c.IsGroup
		}
		g = a.groups[id]
		a.groups[id] = g
	}
	return g
}

// AwayText is the AFK reply: "💤 AFK: reason".
func AwayText(w *Away) string {
	if w.Reason == "" {
		return "💤 AFK"
	}
	return "💤 AFK: " + w.Reason
}

// AwayPlain is the AFK reply without its formatting, as notes show it.
func AwayPlain(w *Away) string {
	return AwayText(w)
}

// cameBack ends AFK because you sent a message, with a notice saying so.
func (a *Backend) cameBack() {
	a.out = append(a.out, model.NoticeEvent{Text: BackText(a.Back())})
}

// BackText says you're no longer AFK, and how many got the reply.
func BackText(told int) string {
	switch told {
	case 0:
		return "Welcome back! You're no longer AFK."
	case 1:
		return "Welcome back! The AFK reply went to 1 person."
	}
	return "Welcome back! The AFK reply went to " + strconv.Itoa(told) + " people."
}

// sent ends AFK when you send something yourself.
func (a *Backend) sent() {
	if a.away != nil {
		a.cameBack()
		if a.notify != nil {
			a.notify() // the notice shows on the next Poll
		}
	}
}

// Ghost reports whether ghost mode is on.
func (a *Backend) Ghost() bool { return a.ghost }

// SetPref keeps a pref, and notices ghost mode turning on or off.
func (a *Backend) SetPref(key, value string) {
	a.Backend.SetPref(key, value)
	if key == model.PrefGhost {
		a.ghost = value == "on"
	}
}

// GhostText is the notice of something refused in ghost mode.
const GhostText = "Ghost mode is on. Turn it off to send messages."

// refused reports whether ghost mode refuses what you're sending, and
// says so.
func (a *Backend) refused() bool {
	if !a.ghost {
		return false
	}
	a.out = append(a.out, model.NoticeEvent{Text: GhostText})
	if a.notify != nil {
		a.notify() // the notice shows on the next Poll
	}
	return true
}

// The ways you send a message, which end AFK, and which ghost mode
// refuses.

func (a *Backend) Send(chatID string, d model.Draft) *model.Message {
	if a.refused() {
		return nil
	}
	a.sent()
	return a.Backend.Send(chatID, d)
}

func (a *Backend) SendSnippet(chatID string, id int64, reply *model.Message, vars map[string]string) (*model.Message, error) {
	if a.refused() {
		return nil, errors.New(GhostText)
	}
	m, err := a.Backend.SendSnippet(chatID, id, reply, vars)
	if err == nil && m != nil {
		a.sent()
	}
	return m, err
}

func (a *Backend) SendFile(chatID string, at model.Attachment, d model.Draft) *model.Message {
	if a.refused() {
		return nil
	}
	a.sent()
	return a.Backend.SendFile(chatID, at, d)
}

func (a *Backend) NewAlbum(chatID string, photos, videos int) string {
	if a.ghost {
		return "" // each picture is refused
	}
	return a.Backend.NewAlbum(chatID, photos, videos)
}

func (a *Backend) SendSticker(chatID string, sticker, reply *model.Message) {
	if a.refused() {
		return
	}
	a.sent()
	a.Backend.SendSticker(chatID, sticker, reply)
}

func (a *Backend) SendNewSticker(chatID string, webp []byte, reply *model.Message) *model.Message {
	if a.refused() {
		return nil
	}
	a.sent()
	return a.Backend.SendNewSticker(chatID, webp, reply)
}

func (a *Backend) SendContacts(chatID string, contactIDs []string) *model.Message {
	if a.refused() {
		return nil
	}
	a.sent()
	return a.Backend.SendContacts(chatID, contactIDs)
}

func (a *Backend) SendPoll(chatID string, p model.Poll) *model.Message {
	if a.refused() {
		return nil
	}
	a.sent()
	return a.Backend.SendPoll(chatID, p)
}

func (a *Backend) Forward(msgs []*model.Message, chatIDs []string) {
	if a.refused() {
		return
	}
	a.sent()
	a.Backend.Forward(msgs, chatIDs)
}

func (a *Backend) PressButton(m *model.Message, i int) *model.Message {
	if a.refused() {
		return nil
	}
	return a.Backend.PressButton(m, i)
}

// What else others would see, which ghost mode refuses too.

func (a *Backend) React(m *model.Message, emoji string) {
	if !a.refused() {
		a.Backend.React(m, emoji)
	}
}

func (a *Backend) VotePoll(m *model.Message, options []int) {
	if !a.refused() {
		a.Backend.VotePoll(m, options)
	}
}

func (a *Backend) Edit(m *model.Message, d model.Draft) {
	if !a.refused() {
		a.Backend.Edit(m, d)
	}
}

// Delete deletes for you in ghost mode too; only for everyone is refused.
func (a *Backend) Delete(m *model.Message, forEveryone bool) {
	if !forEveryone || !a.refused() {
		a.Backend.Delete(m, forEveryone)
	}
}

func (a *Backend) PinMessage(m *model.Message, pinned bool) {
	if !a.refused() {
		a.Backend.PinMessage(m, pinned)
	}
}

func (a *Backend) PostStatus(p model.StatusPost) {
	if !a.refused() {
		a.Backend.PostStatus(p)
	}
}
