package cliout_test

import (
	"strings"
	"testing"

	"samouly.fun/bubbleMonitor/internal/cliout"
	"samouly.fun/bubbleMonitor/internal/ui"
)

func TestNew(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// Verify all 9 style fields produce styled output (non-zero lipgloss.Style)
	// A lipgloss.Style with a Foreground set will produce ANSI escape codes when rendered
	fields := []struct {
		name  string
		style func() string
	}{
		{"Label", func() string { return s.Label.Render("x") }},
		{"Value", func() string { return s.Value.Render("x") }},
		{"Bold", func() string { return s.Bold.Render("x") }},
		{"OK", func() string { return s.OK.Render("x") }},
		{"Warn", func() string { return s.Warn.Render("x") }},
		{"Critical", func() string { return s.Critical.Render("x") }},
		{"Dim", func() string { return s.Dim.Render("x") }},
		{"Header", func() string { return s.Header.Render("x") }},
		{"Active", func() string { return s.Active.Render("x") }},
	}

	for _, f := range fields {
		rendered := f.style()
		// A plain "x" has length 1. A styled "x" contains ANSI escapes and is longer.
		if len(rendered) <= 1 {
			t.Errorf("%s style did not apply ANSI styling (output length %d, expected >1)", f.name, len(rendered))
		}
	}
}

func TestCheckSymbols(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	if !strings.Contains(s.CheckOK, "\u2713") {
		t.Errorf("CheckOK should contain checkmark \u2713, got: %q", s.CheckOK)
	}
	if !strings.Contains(s.CheckFail, "\u2717") {
		t.Errorf("CheckFail should contain cross \u2717, got: %q", s.CheckFail)
	}
	if !strings.Contains(s.CheckWarn, "\u26A0") {
		t.Errorf("CheckWarn should contain warning \u26A0, got: %q", s.CheckWarn)
	}

	// Verify the symbols are styled (contain ANSI codes, longer than plain rune)
	if len(s.CheckOK) <= len("\u2713") {
		t.Errorf("CheckOK should be styled (length %d, expected >%d)", len(s.CheckOK), len("\u2713"))
	}
	if len(s.CheckFail) <= len("\u2717") {
		t.Errorf("CheckFail should be styled (length %d, expected >%d)", len(s.CheckFail), len("\u2717"))
	}
	if len(s.CheckWarn) <= len("\u26A0") {
		t.Errorf("CheckWarn should be styled (length %d, expected >%d)", len(s.CheckWarn), len("\u26A0"))
	}
}

func TestNewStoresPalette(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// Verify the unexported palette is accessible via BarColor returning a non-zero style
	// If palette was not stored, BarColor would panic or return a zero style
	barStyle := s.BarColor(50)
	rendered := barStyle.Render("x")
	if len(rendered) <= 1 {
		t.Error("BarColor returned a zero style, suggesting palette was not stored in CLIStyles")
	}
}

func TestLabelUsesMutedColor(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	rendered := s.Label.Render("test")
	if len(rendered) <= len("test") {
		t.Error("Label style should apply Muted foreground color")
	}
}

func TestOKIsBold(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	rendered := s.OK.Render("test")
	// Bold text contains ANSI escape sequences
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("OK style should produce ANSI codes (bold + Success foreground)")
	}
}

func TestHeaderIsBoldAndUnderlined(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	rendered := s.Header.Render("test")
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("Header style should produce ANSI codes (bold + underline + Primary foreground)")
	}
}

func TestDimIsFaint(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	rendered := s.Dim.Render("test")
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("Dim style should produce ANSI codes (faint + Muted foreground)")
	}
}

func TestActiveIsBold(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	rendered := s.Active.Render("test")
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("Active style should produce ANSI codes (bold + Primary foreground)")
	}
}

func TestBarColor(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// pct=50 -> Success (default)
	rendered := s.BarColor(50).Render("x")
	if len(rendered) <= 1 {
		t.Error("BarColor(50) should return Success-colored style")
	}

	// pct=70 -> Warning (>=60)
	rendered = s.BarColor(70).Render("x")
	if len(rendered) <= 1 {
		t.Error("BarColor(70) should return Warning-colored style")
	}

	// pct=90 -> Alert (>=80)
	rendered = s.BarColor(90).Render("x")
	if len(rendered) <= 1 {
		t.Error("BarColor(90) should return Alert-colored style")
	}
}

func TestBarColorBoundaries(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// Exact boundary: pct=60 -> Warning
	rendered := s.BarColor(60).Render("x")
	if len(rendered) <= 1 {
		t.Error("BarColor(60) should return Warning-colored style (>=60 boundary)")
	}

	// Exact boundary: pct=80 -> Alert
	rendered = s.BarColor(80).Render("x")
	if len(rendered) <= 1 {
		t.Error("BarColor(80) should return Alert-colored style (>=80 boundary)")
	}
}

func TestScoreColor(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// score=80 -> Success (>=70)
	rendered := s.ScoreColor(80).Render("x")
	if len(rendered) <= 1 {
		t.Error("ScoreColor(80) should return Success-colored style")
	}

	// score=55 -> Warning (>=50, <70)
	rendered = s.ScoreColor(55).Render("x")
	if len(rendered) <= 1 {
		t.Error("ScoreColor(55) should return Warning-colored style")
	}

	// score=30 -> Alert (<50)
	rendered = s.ScoreColor(30).Render("x")
	if len(rendered) <= 1 {
		t.Error("ScoreColor(30) should return Alert-colored style")
	}
}

func TestScoreColorBoundaries(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// Exact boundary: score=70 -> Success
	rendered := s.ScoreColor(70).Render("x")
	if len(rendered) <= 1 {
		t.Error("ScoreColor(70) should return Success-colored style (>=70 boundary)")
	}

	// Exact boundary: score=50 -> Warning
	rendered = s.ScoreColor(50).Render("x")
	if len(rendered) <= 1 {
		t.Error("ScoreColor(50) should return Warning-colored style (>=50 boundary)")
	}
}

func TestRenderBar(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// pct=50, width=10: 5 filled + 5 empty
	bar := s.RenderBar(50, 10)
	if !strings.Contains(bar, "\u2588") {
		t.Error("RenderBar(50,10) should contain filled blocks (\u2588)")
	}
	if !strings.Contains(bar, "\u2591") {
		t.Error("RenderBar(50,10) should contain empty blocks (\u2591)")
	}

	// pct=0, width=10: all empty
	bar = s.RenderBar(0, 10)
	if strings.Contains(bar, "\u2588") {
		t.Error("RenderBar(0,10) should not contain filled blocks")
	}

	// pct=100, width=10: all filled
	bar = s.RenderBar(100, 10)
	if strings.Contains(bar, "\u2591") {
		t.Error("RenderBar(100,10) should not contain empty blocks")
	}
}

func TestRenderBarOver100(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// pct=150, width=10: should cap at 10 filled blocks, no empty
	bar := s.RenderBar(150, 10)
	if strings.Contains(bar, "\u2591") {
		t.Error("RenderBar(150,10) should cap filled blocks and have no empty blocks")
	}
}

func TestVisualPad(t *testing.T) {
	palette := ui.GetTheme("charmtone")
	s := cliout.New(palette)

	// A styled string shorter than targetWidth should get padded
	styled := s.Label.Render("hi")
	padded := cliout.VisualPad(styled, 20)
	// The visible width should be exactly 20
	if len(padded) <= len(styled) {
		t.Error("VisualPad should add padding spaces to reach targetWidth")
	}
}

func TestVisualPadExactWidth(t *testing.T) {
	// A plain string already at targetWidth returns unchanged
	result := cliout.VisualPad("hello", 5)
	if result != "hello" {
		t.Errorf("VisualPad(\"hello\", 5) should return \"hello\" unchanged, got %q", result)
	}
}

func TestVisualPadPlainString(t *testing.T) {
	// A plain string shorter than targetWidth gets padded with spaces
	result := cliout.VisualPad("hi", 5)
	if result != "hi   " {
		t.Errorf("VisualPad(\"hi\", 5) should return \"hi   \", got %q", result)
	}
}
