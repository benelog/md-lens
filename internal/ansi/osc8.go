package ansi

// Esc is the ASCII escape character.
const esc = "\x1b"

// stLink is the OSC string terminator (ESC \).
const stLink = esc + "\\"

// Osc8Link builds an OSC 8 terminal hyperlink: ESC]8;;URL ST label ESC]8;; ST.
func Osc8Link(url, label string) string {
	return esc + "]8;;" + url + stLink + label + esc + "]8;;" + stLink
}
