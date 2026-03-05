package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/LFroesch/zap/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	// Help mode
	if m.mode == ModeHelp {
		return ui.HelpScreen(m.width, m.height)
	}

	// Build header
	header := m.renderHeader()

	// Status bar
	statusBar := m.renderStatusBar()

	// Main content
	var mainContent string
	if len(m.configs) == 0 {
		mainContent = m.renderEmptyState()
	} else {
		mainContent = m.renderConfigList()
	}

	// Combine all sections
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainContent,
		statusBar,
	)

	return content
}

func (m model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Inline(true).
		Foreground(lipgloss.Color("214")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(m.width)

	// Create header inside the panel
	sortIcons := []string{"📂", "🕐", "🔤", "📄", "📁"}
	sortNames := []string{"project", "recent", "name", "type", "path"}

	var searchIndicator string
	if m.searchQuery != "" {
		searchIndicator = " [searching]"
	}

	title := "⚡ zap - files registry" + fmt.Sprintf(" [%s %s]%s", sortIcons[m.sortMode], sortNames[m.sortMode], searchIndicator)

	return titleStyle.Render(title)
}

func (m model) renderEmptyState() string {
	availableHeight := m.height - uiOverhead
	if availableHeight < 3 {
		availableHeight = 3
	}

	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(1, 0)

	emptyContent := emptyStyle.Render("📋 No files registered yet.\n\n💡 Press 'n' to add your first file!")

	// Combine header and content
	combined := emptyContent

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 2).
		Height(availableHeight + 2)

	return borderStyle.Render(combined)
}

func (m model) renderConfigList() string {
	// Calculate available height
	availableHeight := m.height - uiOverhead
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Reserve space for scroll indicators (2 lines)
	contentHeight := availableHeight - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Check if we need scroll indicators
	hasTopIndicator := m.scrollOffset > 0
	hasBottomIndicator := m.scrollOffset+contentHeight < len(m.displayConfigs)

	// Adjust for indicators
	actualMaxItems := contentHeight
	if hasTopIndicator {
		actualMaxItems--
	}
	if hasBottomIndicator {
		actualMaxItems--
	}
	if actualMaxItems < 1 {
		actualMaxItems = 1
	}

	// Adjust scroll offset to keep cursor visible
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+actualMaxItems {
		m.scrollOffset = m.cursor - actualMaxItems + 1
	}

	// Recalculate scroll indicators
	hasTopIndicator = m.scrollOffset > 0
	hasBottomIndicator = m.scrollOffset+actualMaxItems < len(m.displayConfigs)

	listStyle := lipgloss.NewStyle().
		Padding(0, 1)

	var items []string

	// Add top scroll indicator
	if hasTopIndicator {
		items = append(items, "▲ more files above...")
	}

	// Calculate visible range
	startIdx := m.scrollOffset
	endIdx := m.scrollOffset + actualMaxItems
	if endIdx > len(m.displayConfigs) {
		endIdx = len(m.displayConfigs)
	}

	// Render visible items
	for i := startIdx; i < endIdx && i < len(m.displayConfigs); i++ {
		display := m.displayConfigs[i]

		var line string
		if display.isHeader {
			// Render header
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))
			line = headerStyle.Render(display.headerText)
		} else {
			// Render config entry
			config := display.config
			name := config.Name
			name = "- " + name

			// Build line: name | project | type | path
			project := config.Project
			if project == "" {
				project = "General"
			}

			// Format with columns
			nameCol := truncate(name, 25)
			projectCol := truncate(project, 15)
			typeCol := truncate(config.Type, 10)
			pathCol := truncate(config.Path, m.width-60)

			line = fmt.Sprintf("%-25s  %-15s  %-10s  %s", nameCol, projectCol, typeCol, pathCol)

			// Apply selection style
			if i == m.cursor {
				selectedStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("214")).
					Foreground(lipgloss.Color("230"))
				line = selectedStyle.Render(line)
			} else {
				normalStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("252"))
				line = normalStyle.Render(line)
			}
		}

		items = append(items, line)
	}

	// Add bottom scroll indicator
	if hasBottomIndicator {
		items = append(items, "▼ more files below...")
	}

	fileList := listStyle.Render(strings.Join(items, "\n"))

	// Combine header and file list
	combined := fileList

	// Combine with border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 2).
		Height(availableHeight + 2)

	return borderStyle.Render(combined)
}

func (m model) renderStatusBar() string {
	// Container style for status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(m.width)

	// Inline styles for colored text (like scout)
	orangeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Background(lipgloss.Color("235")).
		Bold(true).
		Inline(true)

	whiteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("235")).
		Inline(true)

	var statusText string
	var rightSide string

	// Mode-specific status
	switch m.mode {
	case ModeEdit, ModeAdd:
		colNames := []string{"Name", "Project", "Type", "Path", "Description"}
		colName := colNames[m.editCol]

		prefix := "✏️  Editing"
		if m.mode == ModeAdd {
			prefix = "➕ Adding"
		}

		statusText = orangeStyle.Render(prefix+" "+colName+": ") + whiteStyle.Render(m.textInput.View())
		rightSide = orangeStyle.Render("tab") + whiteStyle.Render(": next | ") +
			orangeStyle.Render("enter") + whiteStyle.Render(": save | ") +
			orangeStyle.Render("esc") + whiteStyle.Render(": cancel")

	case ModeSearch:
		matchCount := m.getFilteredConfigsCount()
		statusText = orangeStyle.Render("🔍 Search: ") + whiteStyle.Render(m.searchInput.View())
		rightSide = whiteStyle.Render(fmt.Sprintf("%d matches | ", matchCount)) +
			orangeStyle.Render("enter") + whiteStyle.Render(": apply | ") +
			orangeStyle.Render("esc") + whiteStyle.Render(": cancel")

	case ModeConfirmDelete:
		statusText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("235")).
			Bold(true).
			Inline(true).
			Render(fmt.Sprintf("🗑️  Delete '%s'? ", m.configs[m.deleteIndex].Name))
		rightSide = orangeStyle.Render("y") + whiteStyle.Render(": yes | ") +
			orangeStyle.Render("n/esc") + whiteStyle.Render(": no")

	default:
		// File count
		if len(m.displayConfigs) > 0 {
			// Count actual configs (not headers)
			configCount := 0
			for _, d := range m.displayConfigs {
				if !d.isHeader {
					configCount++
				}
			}
			statusText = orangeStyle.Render(fmt.Sprintf("%d", configCount)) + whiteStyle.Render(" files")
		} else {
			statusText = orangeStyle.Render("0") + whiteStyle.Render(" files")
		}

		// Status message
		if m.statusMsg != "" && time.Now().Before(m.statusExpiry) {
			statusText += whiteStyle.Render(" | " + m.statusMsg)
		}

		// Search query indicator
		if m.searchQuery != "" {
			statusText += whiteStyle.Render(" | ") + orangeStyle.Render(fmt.Sprintf("🔍 '%s'", m.searchQuery))
		}

		// Commands on right
		greenStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Background(lipgloss.Color("235")).
			Bold(true).
			Inline(true)

		redStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("235")).
			Bold(true).
			Inline(true)

		rightSide = greenStyle.Render("o") + whiteStyle.Render(": open | ") +
			orangeStyle.Render("e") + whiteStyle.Render(": edit | ") +
			orangeStyle.Render("N") + whiteStyle.Render(": add | ") +
			redStyle.Render("D") + whiteStyle.Render(": del | ") +
			orangeStyle.Render("y") + whiteStyle.Render(": copy | ") +
			orangeStyle.Render("?") + whiteStyle.Render(": help")
	}

	totalWidth := m.width - 2
	padding := totalWidth - lipgloss.Width(statusText) - lipgloss.Width(rightSide) - 3
	if padding < 1 {
		padding = 1
	}
	statusText += whiteStyle.Render(strings.Repeat(" ", padding)) + rightSide

	return statusStyle.Render(statusText)
}

// truncate truncates a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
