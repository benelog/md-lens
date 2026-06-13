package ansi

import (
	"testing"

	"github.com/benelog/md-lens/internal/term"
)

func TestNoneDepthProducesNoEscapes(t *testing.T) {
	a := New(term.None)
	if got := a.Fg(10, 20, 30); got != "" {
		t.Errorf("fg = %q, want empty", got)
	}
	if got := a.Bold(); got != "" {
		t.Errorf("bold = %q, want empty", got)
	}
	if got := a.Reset(); got != "" {
		t.Errorf("reset = %q, want empty", got)
	}
	if a.Enabled() {
		t.Error("Enabled() = true, want false")
	}
}

func TestTruecolorForegroundAndBackground(t *testing.T) {
	a := New(term.Truecolor)
	if got, want := a.Fg(10, 20, 30), esc+"[38;2;10;20;30m"; got != want {
		t.Errorf("fg = %q, want %q", got, want)
	}
	if got, want := a.Bg(1, 2, 3), esc+"[48;2;1;2;3m"; got != want {
		t.Errorf("bg = %q, want %q", got, want)
	}
	if got, want := a.Bold(), esc+"[1m"; got != want {
		t.Errorf("bold = %q, want %q", got, want)
	}
	if got, want := a.Reset(), esc+"[0m"; got != want {
		t.Errorf("reset = %q, want %q", got, want)
	}
}

func TestRgbTo256Cube(t *testing.T) {
	if got := To256(255, 0, 0); got != 196 {
		t.Errorf("To256(255,0,0) = %d, want 196", got)
	}
	grey := To256(128, 128, 128)
	if grey < 232 || grey > 255 {
		t.Errorf("grey index in ramp: %d", grey)
	}
}

func TestRgbTo16NearestColor(t *testing.T) {
	if got := To16(true, 128, 0, 0); got != 31 {
		t.Errorf("To16 dark red = %d, want 31", got)
	}
	if got := To16(true, 255, 0, 0); got != 91 {
		t.Errorf("To16 bright red = %d, want 91", got)
	}
	if got := To16(false, 0, 0, 0); got != 40 {
		t.Errorf("To16 black bg = %d, want 40", got)
	}
}
