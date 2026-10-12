package main

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
	"github.com/latiefahmad/whatsupclients/internal/ui"
)

// demoQR is a made-up pairing payload of realistic length.
const demoQR = "2@Qm9ndXNQYWlyaW5nUmVmZXJlbmNlRm9yU2NyZWVuc2hvdHNPbmx5X19fX19fX19fX19f,ZmFrZU5vaXNlS2V5X19fX19fX19fX19fX19fX19fXw==,ZmFrZUlkZW50aXR5S2V5X19fX19fX19fX19fX19fXw==,ZmFrZUFkdlNlY3JldF9fX19fX19fX19fX19fX19fXw==,1"

// readmeShot is one of the README's screenshots, rendered from the demo
// backend with the chat open and then the scene set up.
type readmeShot struct {
	name string // the file in docs/screenshots
	dark bool
	chat string
	// scene is a ShowOverlay name, "page:<name>" for ShowPage, "info" for
	// the open chat's info panel, "login" for the QR code, or "".
	scene string
	at    image.Point // where a menu opens, in px at the README's scale
}

// readmeShots are every picture in README.md that the tool can render.
// Add a shot here when the README shows a new one. snippets.PNG has no
// scene hook (the Snippets settings page), so it stays handmade.
var readmeShots = []readmeShot{
	{name: "login.png", scene: "login"},
	{name: "reactions.png", chat: "citra", scene: "msgmenu", at: image.Pt(900, 600)},
	{name: "emoji.png", chat: "citra", scene: "emoji"},
	{name: "poll.png", dark: true, chat: "design", scene: "votes"},
	{name: "msginfo.png", dark: true, chat: "citra", scene: "msginfo"},
	{name: "status.png", chat: "citra", scene: "page:status"},
	{name: "channels.png", dark: true, chat: "citra", scene: "page:channels"},
	{name: "communities.png", chat: "citra", scene: "page:communities"},
	{name: "slash.png", dark: true, chat: "work", scene: "slash"},
	{name: "calc.png", chat: "work", scene: "slashcalc"},
	{name: "afk-list.PNG", dark: true, chat: "citra", scene: "afklist"},
	{name: "privacy.png", dark: true, chat: "citra", scene: "privacy"},
}

// renderReadme writes every readmeShot into out, w×h px.
func renderReadme(out string, w, h int, scale float32) error {
	for _, s := range readmeShots {
		u := ui.New(mock.New())
		if s.scene == "login" {
			u.SetConn(model.ConnEvent{State: model.StateQR, QR: demoQR})
		} else {
			u.Start(func() {})
		}
		u.SetDark(s.dark)
		if s.chat != "" {
			u.SelectID(s.chat)
		}
		switch {
		case s.scene == "" || s.scene == "login":
		case strings.HasPrefix(s.scene, "page:"):
			u.ShowPage(strings.TrimPrefix(s.scene, "page:"))
		default:
			u.ShowOverlay(s.scene, s.at.X, s.at.Y)
		}
		img, err := render(u, w, h, scale)
		if err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
		if err := writeShot(out, s.name, img); err != nil {
			return err
		}
	}
	return nil
}

func writeShot(out, name string, img image.Image) error {
	path := filepath.Join(out, name)
	if err := writePNG(path, img); err != nil {
		return err
	}
	fmt.Println("wrote", path)
	return nil
}
