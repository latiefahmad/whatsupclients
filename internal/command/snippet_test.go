package command

import (
	"strings"
	"testing"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// snippetHost is the Host saveSnippet needs: notes are kept, Do runs
// synchronously, the rest is ignored.
type snippetHost struct {
	notes   []*Note
	changed int
	edited  int64
}

func (h *snippetHost) Note(n *Note)                                     { h.notes = append(h.notes, n) }
func (h *snippetHost) Dismiss(*Note)                                    {}
func (h *snippetHost) Group(model.GroupRequest, func(model.GroupEvent)) {}
func (h *snippetHost) Do(work func() func())                            { work()() }
func (h *snippetHost) Copy(string)                                      {}
func (h *snippetHost) Confirm(_, _, _ string, _ bool, run func())       { run() }
func (h *snippetHost) PickImage(func(string))                           {}
func (h *snippetHost) Sent(*model.Message)                              {}
func (h *snippetHost) Draft(string) model.Draft                         { return model.Draft{} }
func (h *snippetHost) SetGhost(bool)                                    {}
func (h *snippetHost) SnippetVars(string, *model.Message) map[string]string {
	return nil
}
func (h *snippetHost) SnippetsChanged()        { h.changed++ }
func (h *snippetHost) EditSnippet(id int64)    { h.edited = id }
func (h *snippetHost) ShowPayload(_, _ string) {}

func TestSnippetSaveText(t *testing.T) {
	b, h := mock.New(), &snippetHost{}
	c := &Context{Input: "/snippet save hello there", Backend: b, Chat: &model.Chat{ID: "rina"}, Host: h}
	if err := saveSnippet(c, "hello there"); err != nil {
		t.Fatalf("saveSnippet: %v", err)
	}
	if len(h.notes) != 1 || !strings.Contains(h.notes[0].Text, "Saved as Snippet 1") {
		t.Fatalf("notes: %+v", h.notes)
	}
	if h.changed != 1 {
		t.Errorf("SnippetsChanged %d times; want once", h.changed)
	}
	if len(h.notes[0].Buttons) != 1 || h.notes[0].Buttons[0].Label != "Edit in Settings" {
		t.Fatalf("buttons: %+v", h.notes[0].Buttons)
	}
	h.notes[0].Buttons[0].Run()
	if h.edited != 1 {
		t.Errorf("edited %d; want EditSnippet(1)", h.edited)
	}
	items, err := b.Snippets()
	if err != nil || len(items) != 1 || items[0].Name != "Snippet 1" {
		t.Fatalf("snippets: %+v, %v", items, err)
	}
	s, err := b.Snippet(items[0].ID)
	if err != nil || s.Body != "hello there" {
		t.Fatalf("saved body: %+v, %v", s, err)
	}
}

func TestSnippetSaveNeedsTextOrReply(t *testing.T) {
	b, h := mock.New(), &snippetHost{}
	c := &Context{Input: "/snippet save", Backend: b, Chat: &model.Chat{ID: "rina"}, Host: h}
	err := saveSnippet(c, "")
	if err == nil || !strings.Contains(err.Error(), "Reply to a message") {
		t.Fatalf("saveSnippet without reply or text: %v", err)
	}
	if items, _ := b.Snippets(); len(items) != 0 {
		t.Fatalf("nothing to save, but stored: %+v", items)
	}
}

func TestSnippetSaveReply(t *testing.T) {
	b, h := mock.New(), &snippetHost{}
	chats := b.Chats()
	if len(chats) == 0 {
		t.Skip("demo backend has no chats")
	}
	msgs := b.Messages(chats[0].ID, 1)
	if len(msgs) == 0 || msgs[0].Text == "" {
		t.Skip("demo backend has no text message")
	}
	reply := msgs[0]
	c := &Context{Input: "/snippet save", Backend: b, Chat: &model.Chat{ID: reply.ChatID}, Reply: reply, Host: h}
	if err := saveSnippet(c, ""); err != nil {
		t.Fatalf("saveSnippet: %v", err)
	}
	items, err := b.Snippets()
	if err != nil || len(items) != 1 || items[0].Name != "Snippet 1" {
		t.Fatalf("snippets: %+v, %v", items, err)
	}
	s, err := b.Snippet(items[0].ID)
	if err != nil || s.Body != reply.Text {
		t.Fatalf("saved body %q, want %q", s.Body, reply.Text)
	}
}
