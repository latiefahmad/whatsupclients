package ui

import (
	"image"
	"time"

	"gioui.org/op"
)

// Hover tooltips for icon-only buttons, like WhatsApp Desktop's: a small
// label under the button after a short dwell.

// tipDelay is how long the pointer rests on a button before its label shows.
const tipDelay = 500 * time.Millisecond

// tip notes that the button key (showing text) is hovered or not. Call it
// every frame; the label itself is drawn by tipBelow.
func (u *UI) tip(gtx C, key, text string, hovered bool) {
	if !hovered || text == "" {
		if u.tipKey == key {
			u.tipKey = ""
		}
		return
	}
	now := gtx.Now
	if u.tipKey != key {
		u.tipKey, u.tipText, u.tipSince = key, text, now
	}
	if d := tipDelay - now.Sub(u.tipSince); d > 0 {
		gtx.Execute(op.InvalidateCmd{At: now.Add(d)})
	}
}

// tipBelow draws key's label under its button: w wide, pill top at y,
// centered on the button. Call it right where the button is laid out, so
// the label paints over what comes after.
func (u *UI) tipBelow(gtx C, key, text string, w, y int, hovered bool) {
	u.tip(gtx, key, text, hovered)
	if u.tipKey != key || gtx.Now.Sub(u.tipSince) < tipDelay {
		return
	}
	p := u.pal
	m := op.Record(gtx.Ops)
	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}
	cgtx.Constraints.Max.X = min(cgtx.Constraints.Max.X, gtx.Dp(280))
	txt := record(cgtx, u.label(11.5, text, p.Text).Layout)
	pad := image.Pt(gtx.Dp(8), gtx.Dp(4))
	pw, ph := txt.size.X+2*pad.X, txt.size.Y+2*pad.Y
	x := (w - pw) / 2
	fillRRect(gtx, image.Rect(x, y, x+pw, y+ph), gtx.Dp(5), p.Popup)
	borderRRect(gtx, image.Rect(x, y, x+pw, y+ph), gtx.Dp(5), p.Popup, p.PopupBorder)
	txt.at(gtx, x+pad.X, y+pad.Y)
	op.Defer(gtx.Ops, m.Stop())
}
