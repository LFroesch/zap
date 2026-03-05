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

	content.WriteString(titleStyle.Render("zap - Help"))
	content.WriteString("\n\n")

	// Navigation
	content.WriteString(headerStyle.Render("Navigation"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("j/k, up/down") + descStyle.Render("Navigate"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("g/G") + descStyle.Render("First/last item"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("ctrl+d/u") + descStyle.Render("Half-page scroll"))
	content.WriteString("\n")

	// Actions
	content.WriteString(headerStyle.Render("Actions"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("enter/o") + descStyle.Render("Open file in editor"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("O") + descStyle.Render("Open parent directory in editor"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("e") + descStyle.Render("Edit file metadata"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("N") + descStyle.Render("Add new file"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("D") + descStyle.Render("Delete file"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("y") + descStyle.Render("Copy path to clipboard"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("r") + descStyle.Render("Refresh list"))
	content.WriteString("\n")

	// Search & Sort
	content.WriteString(headerStyle.Render("Search & Sort"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("/") + descStyle.Render("Search"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("S") + descStyle.Render("Cycle sort mode"))
	content.WriteString("\n")

	// System
	content.WriteString(headerStyle.Render("System"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render(",") + descStyle.Render("Open config"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("?") + descStyle.Render("Show this help"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("q/ctrl+c") + descStyle.Render("Quit"))
	content.WriteString("\n")

	content.WriteString(footerStyle.Render("Press any key to close help"))

	return content.String()
}
