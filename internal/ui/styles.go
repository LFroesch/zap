package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
const (
	ColorPrimary   = "#7C3AED"
	ColorSuccess   = "#10B981"
	ColorWarning   = "#FBBF24"
	ColorDanger    = "#EF4444"
	ColorInfo      = "#60A5FA"
	ColorMuted     = "#6B7280"
	ColorText      = "#E5E7EB"
	ColorTextLight = "#F3F4F6"
	ColorBorder    = "#374151"
	ColorBg        = "#1F2937"
)

// Styles for the application
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDanger)).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInfo)).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	TextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	EditStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			Bold(true)

	DeleteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDanger)).
			Bold(true)

	SearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInfo)).
			Bold(true)
)

// GetStatusStyle returns the appropriate style based on message content
func GetStatusStyle(message string) lipgloss.Style {
	switch {
	case contains(message, "❌", "Failed", "Error", "Not found"):
		return ErrorStyle
	case contains(message, "⚠️", "Warning"):
		return WarningStyle
	case contains(message, "ℹ️", "Info"):
		return InfoStyle
	default:
		return SuccessStyle
	}
}

func contains(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if lipgloss.NewStyle().Render(s) == lipgloss.NewStyle().Render(substr) ||
		   len(s) > 0 && len(substr) > 0 && s[0:min(len(substr), len(s))] == substr[0:min(len(substr), len(s))] {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
