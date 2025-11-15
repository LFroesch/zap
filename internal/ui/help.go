package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpScreen generates the help screen content
func HelpScreen(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#60A5FA")).
		Width(20)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB"))

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true).
		MarginTop(1).
		Align(lipgloss.Center).
		Width(width)

	var content strings.Builder

	content.WriteString(titleStyle.Render("⚡ zap - Help"))
	content.WriteString("\n\n")

	// Navigation
	content.WriteString(headerStyle.Render("📍 Navigation"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("↑ / k") + descStyle.Render("Move up"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("↓ / j") + descStyle.Render("Move down"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("PageUp / Ctrl+u") + descStyle.Render("Page up"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("PageDown / Ctrl+d") + descStyle.Render("Page down"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("Home") + descStyle.Render("Go to top"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("End") + descStyle.Render("Go to bottom"))
	content.WriteString("\n")

	// Actions
	content.WriteString(headerStyle.Render("⚡ Actions"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("space / enter") + descStyle.Render("Open file in editor"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("e") + descStyle.Render("Edit file metadata"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("n / a") + descStyle.Render("Add new file"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("d") + descStyle.Render("Delete file"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("r") + descStyle.Render("Refresh list"))
	content.WriteString("\n")

	// Search & Filter
	content.WriteString(headerStyle.Render("🔍 Search & Filter"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("/") + descStyle.Render("Start search/filter"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("esc") + descStyle.Render("Clear search"))
	content.WriteString("\n")

	// Sorting
	content.WriteString(headerStyle.Render("📊 Sorting"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("s") + descStyle.Render("Cycle sort modes (5 options)"))
	content.WriteString("\n")

	// System
	content.WriteString(headerStyle.Render("⚙️  System"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("?") + descStyle.Render("Show this help"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("q / ctrl+c") + descStyle.Render("Quit"))
	content.WriteString("\n")

	content.WriteString(footerStyle.Render("Press any key to close help"))

	// Center the content vertically
	availableHeight := height - 4
	contentLines := strings.Count(content.String(), "\n")
	if contentLines < availableHeight {
		padding := (availableHeight - contentLines) / 2
		for i := 0; i < padding; i++ {
			content.WriteString("\n")
		}
	}

	return content.String()
}
