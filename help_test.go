package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHelpViewRenders(t *testing.T) {
	m := model{
		width:    100,
		height:   24,
		mode:     ModeHelp,
		sortMode: 0,
	}

	view := m.View()
	if view == "" {
		t.Fatal("expected help view to render content")
	}
	if !strings.Contains(view, "zap - files registry") {
		t.Fatal("expected top header to remain visible in help mode")
	}
	if got := lipgloss.Height(view); got > m.height {
		t.Fatalf("expected help view height <= %d, got %d", m.height, got)
	}
}
