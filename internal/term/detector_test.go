package term

import "testing"

func nullProbe() (int, int, bool) { return 0, 0, false }

func TestKittyTruecolorTerminal(t *testing.T) {
	caps := detect(
		map[string]string{"COLORTERM": "truecolor", "KITTY_WINDOW_ID": "1", "TERM": "xterm-kitty"},
		true, nullProbe, false, false, false, 0, GraphicsAuto)
	if caps.Color != Truecolor {
		t.Errorf("color = %v, want Truecolor", caps.Color)
	}
	if caps.Graphics != Kitty {
		t.Errorf("graphics = %v, want Kitty", caps.Graphics)
	}
}

func TestPlain256ColorFallsBackToHalfBlock(t *testing.T) {
	caps := detect(map[string]string{"TERM": "xterm-256color"},
		true, nullProbe, false, false, false, 0, GraphicsAuto)
	if caps.Color != Ansi256 {
		t.Errorf("color = %v, want Ansi256", caps.Color)
	}
	if caps.Graphics != HalfBlock {
		t.Errorf("graphics = %v, want HalfBlock", caps.Graphics)
	}
}

func TestNotATtyMeansPlain(t *testing.T) {
	caps := detect(map[string]string{"COLORTERM": "truecolor", "KITTY_WINDOW_ID": "1"},
		false, nullProbe, false, false, false, 0, GraphicsAuto)
	if caps.IsTTY {
		t.Error("IsTTY = true, want false")
	}
	if caps.Color != None {
		t.Errorf("color = %v, want None", caps.Color)
	}
	if caps.Graphics != GraphicsNone {
		t.Errorf("graphics = %v, want GraphicsNone", caps.Graphics)
	}
}

func TestNoColorDisablesColorAndPixelProtocols(t *testing.T) {
	caps := detect(map[string]string{"TERM": "xterm-256color"},
		true, nullProbe, false, true, false, 0, GraphicsAuto)
	if caps.Color != None {
		t.Errorf("color = %v, want None", caps.Color)
	}
	if caps.Graphics != GraphicsNone {
		t.Errorf("graphics = %v, want GraphicsNone", caps.Graphics)
	}
}

func TestForcedGraphicsOverridesDetection(t *testing.T) {
	caps := detect(map[string]string{"KITTY_WINDOW_ID": "1", "COLORTERM": "truecolor"},
		true, nullProbe, false, false, false, 0, HalfBlock)
	if caps.Graphics != HalfBlock {
		t.Errorf("graphics = %v, want HalfBlock", caps.Graphics)
	}
}

func TestSizeFromEnvWhenProbeUnavailable(t *testing.T) {
	caps := detect(map[string]string{"TERM": "xterm-256color", "COLUMNS": "100", "LINES": "40"},
		true, nullProbe, false, false, false, 0, GraphicsAuto)
	if caps.Cols != 100 {
		t.Errorf("cols = %d, want 100", caps.Cols)
	}
	if caps.Rows != 40 {
		t.Errorf("rows = %d, want 40", caps.Rows)
	}
}

func TestForcedWidthCapsColumns(t *testing.T) {
	caps := detect(map[string]string{"TERM": "xterm-256color", "COLUMNS": "100"},
		true, nullProbe, false, false, false, 50, GraphicsAuto)
	if caps.Cols != 50 {
		t.Errorf("cols = %d, want 50", caps.Cols)
	}
}
