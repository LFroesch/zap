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
		content := lipgloss.JoinVertical(lipgloss.Left,
			m.renderHeader(),
			m.renderHelpPanel(),
			m.renderStatusBar(),
		)
		return content
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
	sortIcons := []string{"📂", "🕐", "🔤", "📁"}
	sortNames := []string{"project", "recent", "name", "path"}

	var searchIndicator string
	if m.searchQuery != "" {
		searchIndicator = " [searching]"
	}

	title := "⚡ zap - files registry" + fmt.Sprintf(" [%s %s]%s", sortIcons[m.sortMode], sortNames[m.sortMode], searchIndicator)

	return titleStyle.Render(title)
}

func (m model) renderEmptyState() string {
	availableHeight := m.mainContentHeight()

	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(1, 0)

	emptyContent := emptyStyle.Render("📋 No files registered yet.\n\n💡 Press 'n' to add your first file!")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 2).
		Height(availableHeight)

	return borderStyle.Render(emptyContent)
}

func (m model) renderConfigList() string {
	availableHeight := m.mainContentHeight()
	panelHeight := availableHeight - 2
	if panelHeight < 3 {
		panelHeight = 3
	}

	leftWidth := m.width * 38 / 100
	if leftWidth < 18 {
		leftWidth = 18
	}
	rightWidth := m.width - leftWidth - 1
	if rightWidth < 12 {
		rightWidth = 12
		leftWidth = m.width - rightWidth - 1
	}

	leftPanel := m.renderListPanel(leftWidth, panelHeight)
	rightPanel := m.renderDetailsPanel(rightWidth, panelHeight)

	leftStyled := lipgloss.NewStyle().Height(availableHeight).Render(leftPanel)
	rightStyled := lipgloss.NewStyle().Height(availableHeight).Render(rightPanel)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, " ", rightStyled)
}

func (m model) renderHelpPanel() string {
	return ui.HelpPanel(m.width, m.mainContentHeight(), m.helpScroll)
}

func (m model) renderListPanel(width, panelHeight int) string {
	innerWidth := width - 4
	if innerWidth < 12 {
		innerWidth = 12
	}
	var items []string
	items = append(items, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117")).Render("Files"))
	items = append(items, "")

	maxVisible := panelHeight - len(items)
	// Match sb behavior: keep one row available for potential bottom indicator
	// so visible entries never get clipped when indicator appears.
	maxVisible--
	if maxVisible < 1 {
		maxVisible = 1
	}

	totalRows := len(m.displayConfigs)
	startIdx := 0
	if m.cursor >= maxVisible {
		startIdx = m.cursor - maxVisible + 1
	}
	maxStart := totalRows - maxVisible
	if maxStart < 0 {
		maxStart = 0
	}
	if startIdx > maxStart {
		startIdx = maxStart
	}
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + maxVisible
	if endIdx > totalRows {
		endIdx = totalRows
	}

	for i := startIdx; i < endIdx && i < len(m.displayConfigs); i++ {
		display := m.displayConfigs[i]

		if display.isHeader {
			header := truncate(display.headerText, innerWidth)
			line := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(header)
			items = append(items, line)
			continue
		}

		config := display.config
		name := config.Name
		rawLine := name
		rawLine = truncate(rawLine, innerWidth)

		if i == m.cursor {
			items = append(items, lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Width(innerWidth).
				Render(rawLine))
		} else {
			line := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(rawLine)
			items = append(items, lipgloss.NewStyle().Width(innerWidth).Render(line))
		}
	}

	if startIdx > 0 && len(items) > 1 {
		items[1] = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(fmt.Sprintf("▲ %d more", startIdx))
	}
	if endIdx < totalRows {
		items = append(items, lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(fmt.Sprintf("▼ %d more", totalRows-endIdx)))
	}

	for len(items) < panelHeight {
		items = append(items, "")
	}

	panelContent := strings.Join(items[:panelHeight], "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("117")).
		Padding(0, 1).
		Width(width - 2).
		Height(panelHeight).
		Render(panelContent)
}

func (m model) renderDetailsPanel(width, panelHeight int) string {
	contentWidth := width - 4
	if contentWidth < 12 {
		contentWidth = 12
	}

	var panelContent string
	if m.mode == ModeFileEdit {
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117")).Render("Editing") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("  ctrl+s save · ctrl+d del line · esc cancel")
		panelContent = lipgloss.JoinVertical(lipgloss.Left, header, "", m.fileEditArea.View())
	} else {
		m.rightViewport.Width = contentWidth
		m.rightViewport.Height = panelHeight
		panelContent = m.rightViewport.View()
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width - 2).
		Height(panelHeight).
		Render(panelContent)
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
		colNames := []string{"Name", "Project", "Path", "Description"}
		colName := colNames[m.editCol]

		prefix := "✏️  Editing"
		if m.mode == ModeAdd {
			prefix = "➕ Adding"
		}

		statusText = orangeStyle.Render(prefix+" "+colName+": ") + whiteStyle.Render(m.textInput.View())
		rightSide = orangeStyle.Render("tab") + whiteStyle.Render(": next | ") +
			orangeStyle.Render("enter") + whiteStyle.Render(": save | ") +
			orangeStyle.Render("esc") + whiteStyle.Render(": cancel")

	case ModeFileEdit:
		statusText = orangeStyle.Render("Editing file inline: ") + whiteStyle.Render(m.fileEditLabel)
		rightSide = orangeStyle.Render("ctrl+s") + whiteStyle.Render(": save | ") +
			orangeStyle.Render("ctrl+d") + whiteStyle.Render(": del line | ") +
			orangeStyle.Render("esc") + whiteStyle.Render(": cancel")

	case ModeSearch:
		matchCount := m.getFilteredConfigsCount()
		statusText = orangeStyle.Render("🔍 Search: ") + whiteStyle.Render(m.searchInput.View())
		rightSide = whiteStyle.Render(fmt.Sprintf("%d matches | ", matchCount)) +
			orangeStyle.Render("enter") + whiteStyle.Render(": apply | ") +
			orangeStyle.Render("esc") + whiteStyle.Render(": cancel")

	case ModeHelp:
		statusText = orangeStyle.Render("Help")
		if maxScroll := m.maxHelpScroll(); maxScroll > 0 {
			statusText += whiteStyle.Render(fmt.Sprintf(" | line %d/%d", m.helpScroll+1, maxScroll+1))
		}
		rightSide = orangeStyle.Render("j/k") + whiteStyle.Render(": scroll | ") +
			orangeStyle.Render("pgup/pgdn") + whiteStyle.Render(": page | ") +
			orangeStyle.Render("g/G") + whiteStyle.Render(": top/bottom | ") +
			orangeStyle.Render("esc/?/q") + whiteStyle.Render(": close")

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

		if m.statusMsg != "" && time.Now().Before(m.statusExpiry) {
			statusText += whiteStyle.Render(" | " + m.statusMsg)
		}

		if m.searchQuery != "" {
			statusText += whiteStyle.Render(" | ") + orangeStyle.Render(fmt.Sprintf("🔍 '%s'", m.searchQuery))
		}

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
			greenStyle.Render("E") + whiteStyle.Render(": inline | ") +
			orangeStyle.Render("e") + whiteStyle.Render(": meta | ") +
			orangeStyle.Render("J/K") + whiteStyle.Render(": preview | ") +
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
