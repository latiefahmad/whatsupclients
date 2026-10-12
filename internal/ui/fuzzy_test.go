package ui

import (
	"testing"

	"github.com/latiefahmad/whatsupclients/internal/model"
)

func TestNameScore(t *testing.T) {
	for _, c := range []struct {
		name, q string
		want    int
	}{
		{"Gusti Wibowo", "gusti wibowo", matchExact},
		{"Gusti Wibowo", "GU", matchStart},
		{"~naufalll", "nauf", matchStart},
		{"Gusti Wibowo", "wib", matchWord},
		{"Gusti Wibowo", "wibowo gu", matchWords},
		{"Gusti Wibowo", "ibo", matchInside},
		{"Gusti Wibowo", "gwib", matchLetters}, // starts of words, in order
		{"Gusti Wibowo", "gw", matchLetters},   // initials
		{"Muhammad Salim", "msal", matchLetters},
		{"Gusti Wibowo", "gustu", matchTypo},
		{"Muhammad", "muhamad", matchTypo},
		{"Rizki", "rziki", matchTypo}, // swapped letters
		{"Gusti Wibowo", "ad", 0},     // not at a word's start
		{"Gusti Wibowo", "xyz", 0},
		{"Andi", "bob", 0},
	} {
		if got := nameScore(c.name, foldRunes(c.q)); got != c.want {
			t.Errorf("nameScore(%q, %q) = %d, want %d", c.name, c.q, got, c.want)
		}
	}
}

func TestPhoneScore(t *testing.T) {
	const phone = "+62 857-1111-2222"
	for _, c := range []struct {
		q    string
		want int
	}{
		{"6285711112222", matchExact},
		{"+62 857-1111-2222", matchExact},
		{"62857", matchStart},
		{"+62 857 11", matchStart},
		{"0857", matchWord}, // the national number, with its trunk 0
		{"857-111", matchWord},
		{"2222", matchInside},
		{"1112", matchInside},
		{"22", 0}, // too few digits to look inside
		{"0899", 0},
		{"gusti", 0},
		{"857a", 0},
	} {
		if got := phoneScore(phone, c.q); got != c.want {
			t.Errorf("phoneScore(%q) = %d, want %d", c.q, got, c.want)
		}
	}
	if got := phoneScore("+62 8•• •••• ••90", "890"); got != 0 {
		t.Errorf("a redacted number matched: %d", got)
	}
}

func TestMemberTitle(t *testing.T) {
	for _, c := range []struct {
		m          model.Member
		title, sub string
	}{
		{model.Member{Name: "Gusti", Contact: "Gusti", Push: "gus", Phone: "+62 813-5550-1942"}, "Gusti", ""},
		{model.Member{Name: "+62 857-1111-2222", Push: "naufalll", Phone: "+62 857-1111-2222"}, "~naufalll", "+62 857-1111-2222"},
		{model.Member{Name: "+62 857-1111-2222", Phone: "+62 857-1111-2222"}, "+62 857-1111-2222", ""},
		{model.Member{Name: "~Hidden", Push: "Hidden"}, "~Hidden", ""},
		{model.Member{Name: "Old backend"}, "Old backend", ""},
	} {
		if title, sub := memberTitle(c.m), memberSub(c.m); title != c.title || sub != c.sub {
			t.Errorf("%+v: %q over %q, want %q over %q", c.m, title, sub, c.title, c.sub)
		}
	}
}
