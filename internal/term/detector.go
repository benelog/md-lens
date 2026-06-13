package term

import (
	"os"
	"strconv"
	"strings"

	xterm "golang.org/x/term"
)

// Default cell pixel size when the terminal does not report one (≈2:1, typical monospace).
const (
	defaultCellPxW = 10
	defaultCellPxH = 20
)

// SizeProbe probes the terminal size; returns rows, cols and ok=false if unknown.
type SizeProbe func() (rows, cols int, ok bool)

// Detect detects capabilities against the live process environment.
func Detect(plain, noColor, noImages bool, forcedWidth int, forced GraphicsMode) Capabilities {
	fd := int(os.Stdout.Fd())
	isTty := xterm.IsTerminal(fd)
	probe := func() (int, int, bool) {
		w, h, err := xterm.GetSize(fd)
		if err == nil && w > 0 && h > 0 {
			return h, w, true // rows, cols
		}
		return 0, 0, false
	}
	return detect(envMap(), isTty, probe, plain, noColor, noImages, forcedWidth, forced)
}

// detect holds the pure decision logic so it can be unit-tested with an injected env map and
// size probe — no real terminal needed.
func detect(
	env map[string]string,
	isTty bool,
	probe SizeProbe,
	plain, noColor, noImages bool,
	forcedWidth int,
	forced GraphicsMode,
) Capabilities {
	// Size first — always useful, even when piping.
	rows, cols := 24, 80
	gotSize := false
	if probe != nil {
		if r, c, ok := probe(); ok && r > 0 && c > 0 {
			rows, cols = r, c
			gotSize = true
		}
	}
	if !gotSize {
		cols = parseInt(env["COLUMNS"], cols)
		rows = parseInt(env["LINES"], rows)
	}
	if forcedWidth > 0 {
		cols = forcedWidth
	}

	// Piped / plain output → strip everything so the bytes are clean.
	if !isTty || plain {
		return Capabilities{
			IsTTY: false, Cols: cols, Rows: rows,
			CellPxW: defaultCellPxW, CellPxH: defaultCellPxH,
			Color: None, Graphics: GraphicsNone,
		}
	}

	color := detectColor(env, noColor)
	graphics := GraphicsNone
	if !noImages {
		graphics = detectGraphics(env, forced, color)
	}

	return Capabilities{
		IsTTY: true, Cols: cols, Rows: rows,
		CellPxW: defaultCellPxW, CellPxH: defaultCellPxH,
		Color: color, Graphics: graphics,
	}
}

func detectColor(env map[string]string, noColor bool) ColorDepth {
	if noColor || has(env, "NO_COLOR") {
		return None
	}
	colorterm := lower(env["COLORTERM"])
	if strings.Contains(colorterm, "truecolor") || strings.Contains(colorterm, "24bit") {
		return Truecolor
	}
	t := lower(env["TERM"])
	if t == "" || t == "dumb" {
		return None
	}
	if strings.Contains(t, "256color") || strings.Contains(t, "truecolor") {
		return Ansi256
	}
	return Ansi16
}

func detectGraphics(env map[string]string, forced GraphicsMode, color ColorDepth) GraphicsMode {
	if forced != GraphicsAuto {
		return forced
	}
	t := lower(env["TERM"])
	termProgram := lower(env["TERM_PROGRAM"])

	if has(env, "KITTY_WINDOW_ID") || strings.Contains(t, "kitty") ||
		has(env, "GHOSTTY_RESOURCES_DIR") || strings.Contains(t, "ghostty") {
		return Kitty
	}
	if has(env, "WEZTERM_PANE") || strings.Contains(termProgram, "wezterm") {
		// WezTerm speaks both; kitty protocol is the richer one.
		return Kitty
	}
	if strings.Contains(termProgram, "iterm") || has(env, "ITERM_SESSION_ID") {
		return Iterm2
	}
	// No known pixel protocol: fall back to half-blocks if we have enough colors.
	if color == Truecolor || color == Ansi256 {
		return HalfBlock
	}
	return GraphicsNone
}

func envMap() map[string]string {
	environ := os.Environ()
	m := make(map[string]string, len(environ))
	for _, kv := range environ {
		if i := strings.IndexByte(kv, '='); i >= 0 {
			m[kv[:i]] = kv[i+1:]
		}
	}
	return m
}

func has(env map[string]string, key string) bool {
	_, ok := env[key]
	return ok
}

func lower(s string) string {
	return strings.ToLower(s)
}

func parseInt(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
