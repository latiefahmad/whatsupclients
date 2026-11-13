package ui

import (
	"testing"

	"github.com/latiefahmad/whatsupclients/internal/mock"
	"github.com/latiefahmad/whatsupclients/internal/model"
)

// The demo data's Lottie sticker renders its shapes: decodeScaled gives a
// first frame with pixels in it, and the message says it's a Lottie one.
func TestDemoLottieSticker(t *testing.T) {
	b := mock.New()
	var lottie, plain int
	for _, m := range b.Messages("work", 100) {
		if m.Media != model.MediaSticker {
			continue
		}
		data := b.MediaData(m.ChatID, m.ID)
		if len(data) == 0 {
			continue
		}
		if m.Lottie {
			lottie++
			img, animated := decodeScaled(data, 128)
			if img == nil || !animated {
				t.Fatalf("the Lottie sticker decodes to %v, animated %v", img, animated)
			}
			bounds := img.Bounds()
			opaque := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
						opaque++
					}
				}
			}
			if opaque == 0 {
				t.Error("its first frame drew nothing")
			}
			t.Logf("Lottie sticker first frame: %d opaque px of %d", opaque, bounds.Dx()*bounds.Dy())
		} else {
			plain++
		}
	}
	if lottie != 1 || plain != 0 {
		t.Errorf("%d Lottie and %d plain demo stickers, want 1 and 0", lottie, plain)
	}
}
