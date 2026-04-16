package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func helpLines() []string {
	return []string{
		"Navigation",
		"j/k, up/down        Navigate left file list",
		"g/G                 First/last item",
		"ctrl+d/u            Half-page scroll",
		"J/K                 Scroll right preview pane",
		"pgup/pgdn           Page scroll right preview pane",
		"",
		"Actions",
		"enter/o             Open file in external editor",
		"E                   Edit selected file inline",
		"O                   Open parent directory in editor",
		"e                   Edit selected file metadata",
		"N                   Add new file",
		"D                   Delete file",
		"y                   Copy path to clipboard",
		"r                   Refresh list",
		"",
		"Search & Sort",
		"/                   Search",
		"S                   Cycle sort mode",
		"",
		"Edit Mode",
		"tab/shift+tab       Next/previous field",
		"enter               Save",
		"esc                 Cancel",
		"",
		"Inline File Edit",
		"ctrl+s              Save file",
		"ctrl+d              Delete current line",
		"esc                 Cancel",
		"",
		"System",
		",                   Open config",
		"?                   Show this help",
		"q/ctrl+c            Quit",
	}
}

// HelpLineCount returns the total body line count for the help view.
func HelpLineCount() int {
	return len(helpLines())
}

// HelpPanel renders a bounded help view that preserves the app header/footer.
func HelpPanel(width, height, scroll int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Width(20)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	bodyHeight := height - 4
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	lines := helpLines()
	maxScroll := 0
	if len(lines) > bodyHeight {
		maxScroll = len(lines) - bodyHeight
	}
	if scroll < 0 {
		scroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}

	end := scroll + bodyHeight
	if end > len(lines) {
		end = len(lines)
	}

	var body strings.Builder
	for _, line := range lines[scroll:end] {
		switch {
		case line == "":
			body.WriteString("")
		case !strings.Contains(line, "  "):
			body.WriteString(headerStyle.Render(line))
		default:
			parts := strings.Fields(line)
			key := strings.Join(parts[:1], "")
			desc := strings.TrimSpace(strings.TrimPrefix(line, key))
			if len(parts) > 1 && strings.Contains(line[:min(len(line), 20)], "  ") {
				split := strings.Index(line, "  ")
				if split > 0 {
					key = strings.TrimSpace(line[:split])
					desc = strings.TrimSpace(line[split:])
				}
			}
			body.WriteString(keyStyle.Render(key) + descStyle.Render(desc))
		}
		body.WriteString("\n")
	}

	for visibleLines := end - scroll; visibleLines < bodyHeight; visibleLines++ {
		body.WriteString("\n")
	}

	scrollHint := "j/k scroll • pgup/pgdn page • g/G top/bottom • esc close"
	if maxScroll > 0 {
		scrollHint = fmt.Sprintf("%s • %d/%d", scrollHint, scroll+1, maxScroll+1)
	}

	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Help"),
		"",
		strings.TrimRight(body.String(), "\n"),
		"",
		footerStyle.Render(scrollHint),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("117")).
		Padding(0, 1).
		Width(width - 2).
		Height(height).
		Render(panelContent)
}
