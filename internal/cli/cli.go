// Package cli parses mdl's command-line options.
package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/benelog/md-lens/internal/term"
)

// Args are the parsed command-line options.
type Args struct {
	File            string            // markdown file to render, or "" to read stdin
	NoColor         bool              // disable ANSI color
	NoImages        bool              // do not render images
	Plain           bool              // no styling at all (plain text)
	Width           int               // forced terminal width in columns, or 0 for auto
	ForceGraphics   term.GraphicsMode // override graphics protocol, or GraphicsAuto for auto-detect
	NoHeadingImages bool              // render headings as styled text instead of font images
	Caps            bool              // print detected terminal capabilities and exit
	Help            bool              // show help and exit
	Version         bool              // show version and exit
}

// Parse parses argv into Args. It returns an error for malformed options.
func Parse(argv []string) (Args, error) {
	a := Args{ForceGraphics: term.GraphicsAuto}

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch arg {
		case "--no-color":
			a.NoColor = true
		case "--no-images":
			a.NoImages = true
		case "--plain", "-p":
			a.Plain = true
		case "--no-heading-images":
			a.NoHeadingImages = true
		case "--caps":
			a.Caps = true
		case "--force-kitty":
			a.ForceGraphics = term.Kitty
		case "--force-iterm":
			a.ForceGraphics = term.Iterm2
		case "--force-halfblock":
			a.ForceGraphics = term.HalfBlock
		case "--width", "-w":
			if i+1 >= len(argv) {
				return Args{}, errors.New("--width requires a value")
			}
			i++
			w, err := parseWidth(argv[i])
			if err != nil {
				return Args{}, err
			}
			a.Width = w
		case "--help", "-h":
			a.Help = true
		case "--version", "-V":
			a.Version = true
		default:
			switch {
			case strings.HasPrefix(arg, "--width="):
				w, err := parseWidth(strings.TrimPrefix(arg, "--width="))
				if err != nil {
					return Args{}, err
				}
				a.Width = w
			case strings.HasPrefix(arg, "-") && arg != "-":
				return Args{}, fmt.Errorf("unknown option: %s", arg)
			default:
				a.File = arg
			}
		}
	}
	return a, nil
}

func parseWidth(s string) (int, error) {
	w, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || w <= 0 {
		return 0, fmt.Errorf("invalid --width value: %s", s)
	}
	return w, nil
}
