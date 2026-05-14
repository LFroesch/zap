package suitechrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestJoinHeaderKeepsRightSideVisibleAtNarrowWidth(t *testing.T) {
	left := RenderTitle("zap", "v1.0.0") + " - " + strings.Repeat("files registry ", 3)
	right := Dim("[📂 project] [searching]")

	line := JoinHeader(36, left, right)
	if got := lipgloss.Width(line); got != 36 {
		t.Fatalf("JoinHeader width = %d, want 36", got)
	}
	if !strings.Contains(line, "searching") {
		t.Fatalf("JoinHeader dropped right-side metadata: %q", line)
	}
}

func TestJoinHeaderPadsWhenRightSideEmpty(t *testing.T) {
	line := JoinHeader(20, "zap header", "")
	if got := lipgloss.Width(line); got != 20 {
		t.Fatalf("JoinHeader width = %d, want 20", got)
	}
}
