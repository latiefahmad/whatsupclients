package mock

import (
	"archive/zip"
	"bytes"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

// demoLottie gives one demo sticker a Lottie animation (a WhatsApp .was),
// so the chat shows a sticker that plays from its shapes.
//
// Its JSON is a 512x512 animation of a rounded blob that grows, turns and
// fades over two seconds at 30 a second, which the player draws itself.
func (b *Backend) demoLottie() {
	const chat, id = "work", "work-6"
	var m *model.Message
	for _, msg := range b.msgs[chat] {
		if msg.ID == id && msg.Media == model.MediaSticker {
			m = msg
			break
		}
	}
	if m == nil {
		return
	}
	m.Lottie = true
	b.mu.Lock()
	if b.media == nil {
		b.media = map[string][]byte{}
	}
	b.media[chat+"/"+id] = wasBytes(demoLottieJSON)
	b.mu.Unlock()
}

// demoLottieJSON is the animation demoLottie is made of.
const demoLottieJSON = `{"v": "5.7.0", "fr": 30, "ip": 0, "op": 60, "w": 512, "h": 512,
 "layers": [
  {"ty": 4, "ind": 1, "ip": 0, "op": 60, "st": 0,
   "ks": {"p": {"a": 0, "k": [256, 256]},
          "r": {"a": 1, "k": [{"t": 0, "s": [0]}, {"t": 59, "s": [180]}]},
          "s": {"a": 1, "k": [{"t": 0, "s": [60, 60]}, {"t": 30, "s": [190, 190]}, {"t": 59, "s": [120, 120]}]}},
   "shapes": [
    {"ty": "sr", "sy": 1, "d": 1, "pt": {"a": 0, "k": 5}, "p": {"a": 0, "k": [0, -60]},
     "r": {"a": 0, "k": 0}, "or": {"a": 0, "k": 70}, "ir": {"a": 0, "k": 34}, "os": {"a": 0, "k": 0}},
    {"ty": "fl", "c": {"a": 0, "k": [0.15, 0.72, 0.46, 1]},
     "o": {"a": 1, "k": [{"t": 0, "s": [100]}, {"t": 45, "s": [100]}, {"t": 59, "s": [0]}]}}]},
  {"ty": 4, "ind": 2, "ip": 0, "op": 60, "st": 0,
   "ks": {"p": {"a": 0, "k": [256, 256]},
          "r": {"a": 1, "k": [{"t": 0, "s": [0]}, {"t": 59, "s": [-120]}]},
          "s": {"a": 1, "k": [{"t": 0, "s": [40, 40]}, {"t": 40, "s": [150, 150]}, {"t": 59, "s": [90, 90]}]}},
   "shapes": [
    {"ty": "el", "p": {"a": 0, "k": [0, 0]}, "s": {"a": 0, "k": [200, 200]}},
    {"ty": "fl", "c": {"a": 0, "k": [1, 0.85, 0.25, 1]},
     "o": {"a": 1, "k": [{"t": 0, "s": [100]}, {"t": 50, "s": [100]}, {"t": 59, "s": [0]}]}}]}]}`

// wasBytes packs a Lottie animation as WhatsApp's .was: a zip holding
// animation/animation.json.
func wasBytes(anim string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("animation/animation.json")
	_, _ = w.Write([]byte(anim))
	_ = zw.Close()
	return buf.Bytes()
}
