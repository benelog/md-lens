package ansi

import "testing"

func TestWidthPlainAscii(t *testing.T) {
	if got := Width("abc"); got != 3 {
		t.Errorf("Width(abc) = %d, want 3", got)
	}
}

func TestWidthIgnoresSgrSequences(t *testing.T) {
	styled := esc + "[1m" + esc + "[38;2;1;2;3m" + "abc" + esc + "[0m"
	if got := Width(styled); got != 3 {
		t.Errorf("Width(styled) = %d, want 3", got)
	}
}

func TestWidthCountsEastAsianWideAsTwo(t *testing.T) {
	if got := Width("한글"); got != 4 {
		t.Errorf("Width(한글) = %d, want 4", got)
	}
	if got := Width("한a글a"); got != 6 { // 2 + 1 + 2 + 1
		t.Errorf("Width(한a글a) = %d, want 6", got)
	}
}

func TestWidthIgnoresOsc8Hyperlink(t *testing.T) {
	if got := Width(Osc8Link("https://example.com", "link")); got != 4 {
		t.Errorf("Width(osc8 link) = %d, want 4", got)
	}
}
