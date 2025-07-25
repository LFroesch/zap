package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfigEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`        // json, yaml, toml, ini, txt
	Project     string `json:"project"`     // project association
	Description string `json:"description"` // brief description
}

type ConfigManager struct {
	Configs []ConfigEntry `json:"configs"`
}

type statusMsg struct {
	message string
}

func showStatus(msg string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{message: msg}
	}
}

type model struct {
	configs       []ConfigEntry
	table         table.Model
	editMode      bool
	editRow       int
	editCol       int
	textInput     textinput.Model
	configFile    string
	width         int
	height        int
	statusMsg     string
	statusExpiry  time.Time
	scrollOffset  int            // For horizontal scrolling
	maxCols       int            // Maximum visible columns
	configIndices []int          // Maps display row to actual config index (-1 for headers)
	allColumns    []table.Column // Store all possible columns
	confirmDelete bool           // Confirmation mode for deletion
	deleteIndex   int            // Index of config to delete
	viewMode      bool           // View/edit config content mode
	viewContent   string         // Content of config being viewed
	viewPath      string         // Path of config being viewed
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configFile := filepath.Join(homeDir, ".config", "zap", "zap-registry.json")

	m := model{
		configs:       loadConfigs(configFile),
		configFile:    configFile,
		width:         100,
		height:        24,
		editMode:      false,
		editRow:       -1,
		editCol:       -1,
		scrollOffset:  0,
		maxCols:       5,
		confirmDelete: false,
		deleteIndex:   -1,
		viewMode:      false,
	}

	// Define all possible columns
	m.allColumns = []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Project", Width: 20},
		{Title: "Type", Width: 10},
		{Title: "Path", Width: 40},
		{Title: "Description", Width: 30},
	}

	// Initialize text input for editing
	m.textInput = textinput.New()
	m.textInput.CharLimit = 300

	// Initialize table with initial columns
	t := table.New(
		table.WithColumns(m.allColumns[:4]), // Start with first 4 columns
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#F3F4F6")).
		Background(lipgloss.Color("#1F2937"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#F3F4F6")).
		Background(lipgloss.Color("#7C3AED")).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("#E5E7EB"))
	t.SetStyles(s)

	m.table = t
	m.adjustLayout()
	m.updateTable()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func loadConfigs(configFile string) []ConfigEntry {
	var manager ConfigManager
	data, err := os.ReadFile(configFile)
	if err != nil {
		// Create default config directory
		os.MkdirAll(filepath.Dir(configFile), 0755)
		return []ConfigEntry{}
	}
	json.Unmarshal(data, &manager)
	return manager.Configs
}

func (m *model) saveConfigs() {
	manager := ConfigManager{Configs: m.configs}
	data, err := json.MarshalIndent(manager, "", "  ")
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(m.configFile), 0755)
	os.WriteFile(m.configFile, data, 0644)
}

func (m *model) updateTable() {
	sortedConfigs := m.getSortedConfigs()
	visibleColumns := m.table.Columns()

	var rows []table.Row
	m.configIndices = []int{} // Reset config indices mapping

	var lastProject string
	configIndex := 0

	for _, config := range sortedConfigs {
		// Handle empty project display
		displayProject := config.Project
		if displayProject == "" {
			displayProject = "General"
		}

		// Add project header if this is a new project
		if displayProject != lastProject {
			// Create project header row
			projectHeader := fmt.Sprintf("📂 %s", displayProject)

			// Create header row with same number of columns as visible columns
			headerRow := make(table.Row, len(visibleColumns))
			headerRow[0] = projectHeader // Show project in first column
			for i := 1; i < len(headerRow); i++ {
				headerRow[i] = ""
			}

			rows = append(rows, headerRow)
			m.configIndices = append(m.configIndices, -1) // -1 indicates header row
			lastProject = displayProject
		}

		// Create config row - build full row data first
		fullRowData := []string{config.Name, displayProject, config.Type, config.Path, config.Description}

		// Create visible row based on current visible columns and scroll offset
		visibleRow := make(table.Row, len(visibleColumns))
		for i, col := range visibleColumns {
			columnIndex := m.getColumnIndex(col.Title)
			if columnIndex >= 0 && columnIndex < len(fullRowData) {
				visibleRow[i] = fullRowData[columnIndex]
			} else {
				visibleRow[i] = ""
			}
		}

		rows = append(rows, visibleRow)
		m.configIndices = append(m.configIndices, configIndex)
		configIndex++
	}
	m.table.SetRows(rows)
}

func (m *model) getColumnIndex(title string) int {
	switch title {
	case "Name":
		return 0
	case "Project":
		return 1
	case "Type":
		return 2
	case "Path":
		return 3
	case "Description":
		return 4
	default:
		return -1
	}
}

func (m *model) adjustLayout() {
	tableHeight := m.height - 6
	if tableHeight < 5 {
		tableHeight = 5
	}

	// Calculate available width for columns
	availableWidth := m.width - 6 // Account for borders

	// Calculate how many columns can fit
	totalWidth := 0
	visibleCols := 0
	for i, col := range m.allColumns {
		if totalWidth+col.Width <= availableWidth {
			totalWidth += col.Width
			visibleCols++
		} else {
			break
		}
		if i >= len(m.allColumns)-1 {
			break
		}
	}

	// Ensure we show at least one column
	if visibleCols == 0 {
		visibleCols = 1
		// Create a copy of the first column and adjust width
		firstCol := m.allColumns[0]
		firstCol.Width = availableWidth
		m.allColumns[0] = firstCol
	}

	// Apply horizontal scrolling offset
	startCol := m.scrollOffset
	endCol := startCol + visibleCols
	if endCol > len(m.allColumns) {
		endCol = len(m.allColumns)
		startCol = endCol - visibleCols
		if startCol < 0 {
			startCol = 0
		}
		m.scrollOffset = startCol
	}

	// Select visible columns
	var visibleColumns []table.Column
	for i := startCol; i < endCol && i < len(m.allColumns); i++ {
		visibleColumns = append(visibleColumns, m.allColumns[i])
	}

	// If we have extra space, distribute it among visible columns
	if len(visibleColumns) > 0 {
		usedWidth := 0
		for _, col := range visibleColumns {
			usedWidth += col.Width
		}
		if extraWidth := availableWidth - usedWidth; extraWidth > 0 {
			// Distribute extra width to the last column
			visibleColumns[len(visibleColumns)-1].Width += extraWidth
		}
	}

	m.table.SetColumns(visibleColumns)
	m.table.SetHeight(tableHeight)
	m.maxCols = len(m.allColumns)
}

func (m *model) startEdit() {
	if len(m.configs) == 0 {
		return
	}

	m.editMode = true
	displayIndex := m.table.Cursor()
	m.editRow = m.getOriginalIndexByDisplayIndex(displayIndex)
	if m.editRow == -1 {
		return // Invalid index
	}
	m.editCol = 0 // Start with name column

	// Set the current value in the text input
	config := m.configs[m.editRow]
	var initialValue string
	switch m.editCol {
	case 0:
		initialValue = config.Name
	case 1:
		initialValue = config.Project
	case 2:
		initialValue = config.Type
	case 3:
		initialValue = config.Path
	case 4:
		initialValue = config.Description
	}
	m.textInput.SetValue(initialValue)
	m.textInput.SetCursor(len(initialValue))
	m.textInput.Focus()
}

func (m *model) saveEdit() {
	if !m.editMode || m.editRow < 0 || m.editRow >= len(m.configs) {
		return
	}

	value := m.textInput.Value()
	switch m.editCol {
	case 0:
		m.configs[m.editRow].Name = value
	case 1:
		m.configs[m.editRow].Project = value
	case 2:
		m.configs[m.editRow].Type = value
	case 3:
		m.configs[m.editRow].Path = expandPath(value)
	case 4:
		m.configs[m.editRow].Description = value
	}

	m.saveConfigs()
	m.updateTable()
}

func (m *model) cancelEdit() {
	m.editMode = false
	m.editRow = -1
	m.editCol = -1
	m.textInput.Blur()
	m.textInput.SetValue("")
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("zap - File Registry") // Set initial window title
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case statusMsg:
		m.statusMsg = msg.message
		m.statusExpiry = time.Now().Add(3 * time.Second)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustLayout()
		m.updateTable()
		return m, nil

	case tea.KeyMsg:
		if m.viewMode {
			return m.updateView(msg)
		}
		if m.editMode {
			return m.updateEdit(msg)
		}
		if m.confirmDelete {
			return m.updateDeleteConfirm(msg)
		}
		return m.updateNormal(msg)
	}

	// Let table handle mouse events when not editing
	if !m.editMode && !m.viewMode {
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewMode = false
		m.viewContent = ""
		m.viewPath = ""
		return m, nil
	case "e":
		// Exit view mode and open in external editor
		m.viewMode = false
		m.viewContent = ""
		displayIndex := m.table.Cursor()
		config := m.getConfigByDisplayIndex(displayIndex)
		if config != nil {
			return m, m.openInEditor(*config)
		}
		return m, nil
	}
	return m, nil
}

func (m model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm deletion
		if m.deleteIndex >= 0 && m.deleteIndex < len(m.configs) {
			configName := m.configs[m.deleteIndex].Name
			m.configs = append(m.configs[:m.deleteIndex], m.configs[m.deleteIndex+1:]...)
			m.saveConfigs()
			m.updateTable()
			m.confirmDelete = false
			m.deleteIndex = -1
			return m, showStatus(fmt.Sprintf("🗑️ Deleted %s", configName))
		}
		m.confirmDelete = false
		m.deleteIndex = -1
		return m, nil
	case "n", "N", "esc":
		// Cancel deletion
		m.confirmDelete = false
		m.deleteIndex = -1
		return m, showStatus("❌ Deletion cancelled")
	}
	return m, nil
}

func (m model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.cancelEdit()
		return m, nil
	case "enter":
		m.saveEdit()
		m.cancelEdit()
		return m, showStatus("✅ File updated")
	case "tab":
		// Save current field and move to next
		m.saveEdit()
		m.editCol = (m.editCol + 1) % 5
		config := m.configs[m.editRow]
		var newValue string
		switch m.editCol {
		case 0:
			newValue = config.Name
		case 1:
			newValue = config.Project
		case 2:
			newValue = config.Type
		case 3:
			newValue = config.Path
		case 4:
			newValue = config.Description
		}
		m.textInput.SetValue(newValue)
		m.textInput.SetCursor(len(newValue))
		return m, nil
	case "shift+tab":
		// Save current field and move to previous
		m.saveEdit()
		m.editCol = (m.editCol - 1 + 5) % 5
		config := m.configs[m.editRow]
		var newValue string
		switch m.editCol {
		case 0:
			newValue = config.Name
		case 1:
			newValue = config.Project
		case 2:
			newValue = config.Type
		case 3:
			newValue = config.Path
		case 4:
			newValue = config.Description
		}
		m.textInput.SetValue(newValue)
		m.textInput.SetCursor(len(newValue))
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "e":
		m.startEdit()
		return m, nil
	case "n", "a":
		// Add new config
		newConfig := ConfigEntry{
			Name:        "New File",
			Path:        "~/path/to/file.json",
			Type:        "json",
			Project:     "",
			Description: "File description",
		}
		m.configs = append(m.configs, newConfig)
		m.saveConfigs()
		m.updateTable()
		// Find the display index of the newly added config
		displayIndex := m.findConfigDisplayIndex(newConfig)
		if displayIndex != -1 {
			m.table.SetCursor(displayIndex)
			m.startEdit()
		}
		return m, showStatus("➕ New file added")
	case "d", "delete":
		if len(m.configs) > 0 {
			displayIndex := m.table.Cursor()
			originalIndex := m.getOriginalIndexByDisplayIndex(displayIndex)
			if originalIndex == -1 {
				return m, nil
			}
			m.confirmDelete = true
			m.deleteIndex = originalIndex
			return m, showStatus(fmt.Sprintf("❓ Delete '%s'? (y/n)", m.configs[originalIndex].Name))
		}
		return m, nil
	case " ", "enter":
		if len(m.configs) > 0 {
			displayIndex := m.table.Cursor()
			config := m.getConfigByDisplayIndex(displayIndex)
			if config != nil {
				return m, m.openInEditor(*config)
			}
		}
		return m, nil

	case "r":
		m.configs = loadConfigs(m.configFile)
		m.updateTable()
		return m, showStatus("🔄 Refreshed")
	case "left":
		// Horizontal scroll left
		if m.scrollOffset > 0 {
			m.scrollOffset--
			m.adjustLayout()
			m.updateTable()
		}
		return m, nil
	case "right":
		// Horizontal scroll right
		maxOffset := m.maxCols - len(m.table.Columns())
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.scrollOffset < maxOffset {
			m.scrollOffset++
			m.adjustLayout()
			m.updateTable()
		}
		return m, nil
	default:
		// Let table handle arrow keys and other navigation
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) openInEditor(config ConfigEntry) tea.Cmd {
	return func() tea.Msg {
		// Expand path
		expandedPath := expandPath(config.Path)

		// Check if file exists
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return statusMsg{message: fmt.Sprintf("❌ File not found: %s", expandedPath)}
		}

		// Try different editors in order of preference
		editors := []string{"code", "nano", "vim", "vi"}
		var cmd *exec.Cmd

		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				if editor == "code" {
					// VS Code
					cmd = exec.Command(editor, expandedPath)
				} else {
					// Terminal editors
					cmd = exec.Command(editor, expandedPath)
				}
				break
			}
		}

		if cmd == nil {
			return statusMsg{message: "❌ No suitable editor found (tried: code, nano, vim, vi)"}
		}

		err := cmd.Start()
		if err != nil {
			return statusMsg{message: fmt.Sprintf("❌ Failed to open editor: %v", err)}
		}

		return statusMsg{message: fmt.Sprintf("📝 Opened %s in editor", config.Name)}
	}
}

func (m model) viewConfig(config ConfigEntry) tea.Cmd {
	return func() tea.Msg {
		// Expand path
		expandedPath := expandPath(config.Path)

		// Read file content
		content, err := os.ReadFile(expandedPath)
		if err != nil {
			return statusMsg{message: fmt.Sprintf("❌ Failed to read file: %v", err)}
		}

		// Update model to show content
		m.viewMode = true
		m.viewContent = string(content)
		m.viewPath = expandedPath

		return statusMsg{message: fmt.Sprintf("👁️ Viewing %s (press 'e' to edit, 'esc' to close)", config.Name)}
	}
}

func (m *model) getSortedConfigs() []ConfigEntry {
	// Create a copy of configs for sorting
	sortedConfigs := make([]ConfigEntry, len(m.configs))
	copy(sortedConfigs, m.configs)

	// Sort configs by project first, then by name within each project
	sort.Slice(sortedConfigs, func(i, j int) bool {
		// Handle empty projects by treating them as "General"
		projectI := sortedConfigs[i].Project
		if projectI == "" {
			projectI = "General"
		}
		projectJ := sortedConfigs[j].Project
		if projectJ == "" {
			projectJ = "General"
		}

		// First sort by project
		if !strings.EqualFold(projectI, projectJ) {
			return strings.ToLower(projectI) < strings.ToLower(projectJ)
		}

		// If projects are the same, sort by name
		return strings.ToLower(sortedConfigs[i].Name) < strings.ToLower(sortedConfigs[j].Name)
	})

	return sortedConfigs
}

func (m *model) getConfigByDisplayIndex(displayIndex int) *ConfigEntry {
	// Check if the display index is valid and not a header row
	if displayIndex < 0 || displayIndex >= len(m.configIndices) {
		return nil
	}

	// Get the actual config index (-1 means header row)
	configIndex := m.configIndices[displayIndex]
	if configIndex == -1 {
		return nil // This is a header row, no config associated
	}

	sortedConfigs := m.getSortedConfigs()
	if configIndex >= len(sortedConfigs) {
		return nil
	}

	// Find the original config in m.configs that matches the sorted config
	sortedConfig := sortedConfigs[configIndex]
	for i := range m.configs {
		if m.configs[i].Name == sortedConfig.Name &&
			m.configs[i].Path == sortedConfig.Path &&
			m.configs[i].Project == sortedConfig.Project {
			return &m.configs[i]
		}
	}
	return nil
}

func (m *model) findConfigDisplayIndex(targetConfig ConfigEntry) int {
	// Find the display index of a config in the table
	for i, configIndex := range m.configIndices {
		if configIndex == -1 {
			continue // Skip header rows
		}

		sortedConfigs := m.getSortedConfigs()
		if configIndex < len(sortedConfigs) {
			config := sortedConfigs[configIndex]
			if config.Name == targetConfig.Name &&
				config.Path == targetConfig.Path &&
				config.Project == targetConfig.Project {
				return i
			}
		}
	}
	return -1
}

func (m *model) getOriginalIndexByDisplayIndex(displayIndex int) int {
	// Check if the display index is valid and not a header row
	if displayIndex < 0 || displayIndex >= len(m.configIndices) {
		return -1
	}

	// Get the actual config index (-1 means header row)
	configIndex := m.configIndices[displayIndex]
	if configIndex == -1 {
		return -1 // This is a header row, no config associated
	}

	sortedConfigs := m.getSortedConfigs()
	if configIndex >= len(sortedConfigs) {
		return -1
	}

	// Find the original index in m.configs that matches the sorted config
	sortedConfig := sortedConfigs[configIndex]
	for i := range m.configs {
		if m.configs[i].Name == sortedConfig.Name &&
			m.configs[i].Path == sortedConfig.Path &&
			m.configs[i].Project == sortedConfig.Project {
			return i
		}
	}
	return -1
}

func (m model) View() string {
	if m.viewMode {
		// View mode - show file content
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

		contentStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(1).
			Height(m.height - 8).
			Width(m.width - 4)

		footer := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Render("e: edit in external editor • esc/q: close view")

		title := titleStyle.Render(fmt.Sprintf("📄 Viewing: %s", m.viewPath))
		content := contentStyle.Render(m.viewContent)

		return lipgloss.JoinVertical(lipgloss.Left, title, "", content, "", footer)
	}

	// Normal table view
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true)
	header := titleStyle.Render("⚡ zap - File Registry")

	if len(m.configs) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1).
			MarginBottom(1)

		content := emptyStyle.Render("📋 No files registered yet.\n\n💡 Press 'n' to add your first file!")
		footer := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			Render("Commands: ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Render("n/a: add file") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(" • ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Render("q: quit")

		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			content,
			footer,
		)
	}

	var statusMessage string
	if m.statusMsg != "" && time.Now().Before(m.statusExpiry) {
		// Color code based on message type
		if strings.Contains(m.statusMsg, "❌") || strings.Contains(m.statusMsg, "Failed") {
			statusMessage = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Bold(true).
				Render("Status: " + m.statusMsg)
		} else {
			statusMessage = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Bold(true).
				Render("Status: " + m.statusMsg)
		}
	}

	// Show different footer based on mode
	var footer string
	if m.editMode {
		colNames := []string{"Name", "Project", "Type", "Path", "Description"}
		colName := colNames[m.editCol]

		typeHelp := ""
		if colName == "Type" {
			typeHelp = " (json, yaml, toml, ini, txt)"
		}

		editStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

		footer = editStyle.Render(fmt.Sprintf("✏️  Editing %s%s: %s", colName, typeHelp, m.textInput.View())) +
			helpStyle.Render("\nCommands: tab: next field • enter: save • esc: cancel")
	} else if m.confirmDelete {
		deleteStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DC2626")).
			Bold(true)

		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

		footer = deleteStyle.Render(fmt.Sprintf("🗑️  Delete '%s'? ", m.configs[m.deleteIndex].Name)) +
			helpStyle.Render("y: yes • n/esc: no")
	} else {
		// Style individual command groups with colors
		navStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))    // Blue
		actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")) // Green
		editStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))   // Yellow
		systemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")) // Red
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))

		scrollHint := ""
		if m.maxCols > len(m.table.Columns()) {
			scrollHint = " • " + navStyle.Render("←→: scroll columns")
		}

		commands := []string{
			navStyle.Render("↑↓: navigate") + scrollHint,
			actionStyle.Render("space/enter: edit"),
			editStyle.Render("e: edit fields"),
			editStyle.Render("n/a: add"),
			systemStyle.Render("d: delete"),
			systemStyle.Render("r: refresh"),
			systemStyle.Render("q: quit"),
		}
		footer = helpStyle.Render("Commands: " + strings.Join(commands, " • "))
	}

	// Build the final view
	var parts []string

	// Always include header
	parts = append(parts, header)

	// Add table
	parts = append(parts, m.table.View())

	// Add status message if present
	if statusMessage != "" {
		parts = append(parts, statusMessage)
	}

	// Add footer
	parts = append(parts, footer)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
