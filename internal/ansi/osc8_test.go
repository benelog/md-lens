package ansi

import "testing"

func TestEmitsHyperlinkSequence(t *testing.T) {
	want := esc + "]8;;https://x" + esc + "\\" + "label" + esc + "]8;;" + esc + "\\"
	if got := Osc8Link("https://x", "label"); got != want {
		t.Errorf("Osc8Link = %q, want %q", got, want)
	}
}
