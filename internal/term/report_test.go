package term

import (
	"strings"
	"testing"
)

func TestHalfBlockTerminalReportsTextHeadings(t *testing.T) {
	caps := Capabilities{IsTTY: true, Cols: 100, Rows: 40, CellPxW: 10, CellPxH: 20,
		Color: Truecolor, Graphics: HalfBlock}
	report := FormatReport(caps, true)
	for _, want := range []string{"truecolor", "half-block", "styled text", "100 x 40"} {
		if !strings.Contains(report, want) {
			t.Errorf("report missing %q:\n%s", want, report)
		}
	}
}

func TestKittyTerminalReportsFontImageHeadings(t *testing.T) {
	caps := Capabilities{IsTTY: true, Cols: 80, Rows: 24, CellPxW: 10, CellPxH: 20,
		Color: Truecolor, Graphics: Kitty}
	report := FormatReport(caps, true)
	for _, want := range []string{"kitty", "large font images"} {
		if !strings.Contains(report, want) {
			t.Errorf("report missing %q:\n%s", want, report)
		}
	}
}

func TestNoHeadingImagesFlagDisablesFontHeadings(t *testing.T) {
	caps := Capabilities{IsTTY: true, Cols: 80, Rows: 24, CellPxW: 10, CellPxH: 20,
		Color: Truecolor, Graphics: Kitty}
	report := FormatReport(caps, false)
	if !strings.Contains(report, "styled text") {
		t.Errorf("report missing %q:\n%s", "styled text", report)
	}
}
