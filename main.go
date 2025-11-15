package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"configly/internal/editor"
	"configly/internal/models"
	"configly/internal/storage"
	"configly/internal/ui"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the different view states
type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeEdit
	ModeAdd
	ModeSearch
	ModeHelp
	ModeConfirmDelete
)

type model struct {
	configs       []models.ConfigEntry
	table         table.Model
	storage       *storage.Storage
	width         int
	height        int

	// Mode management
	mode          ViewMode

	// Edit mode
	editRow       int
	editCol       int
	textInput     textinput.Model

	// Search mode
	searchInput   textinput.Model
	searchQuery   string
	fuzzyMode     bool

	// Delete confirmation
	deleteIndex   int

	// UI state
	statusMsg     string
	statusExpiry  time.Time
	scrollOffset  int
	maxCols       int
	configIndices []int
	allColumns    []table.Column

	// Performance
	sortedCache   []models.ConfigEntry
	cacheValid    bool
	sortByRecent  bool
}

type statusMsg struct {
	message string
}

func showStatus(msg string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{message: msg}
	}
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configFile := filepath.Join(homeDir, ".config", "zap", "zap-registry.json")

	store := storage.New(configFile)
	configs, err := store.Load()
	if err != nil {
		log.Fatalf("Failed to load configs: %v", err)
	}

	m := model{
		configs:      configs,
		storage:      store,
		width:        100,
		height:       24,
		mode:         ModeNormal,
		editRow:      -1,
		editCol:      -1,
		scrollOffset: 0,
		maxCols:      5,
		deleteIndex:  -1,
		cacheValid:   false,
		sortByRecent: false,
	}

	// Define all possible columns
	m.allColumns = []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Project", Width: 20},
		{Title: "Type", Width: 10},
		{Title: "Path", Width: 40},
		{Title: "Description", Width: 30},
	}

	// Initialize text inputs
	m.textInput = textinput.New()
	m.textInput.CharLimit = 300

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Type to search..."
	m.searchInput.CharLimit = 100

	// Initialize table
	t := table.New(
		table.WithColumns(m.allColumns[:4]),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ui.ColorBorder)).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color(ui.ColorTextLight)).
		Background(lipgloss.Color(ui.ColorBg))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(ui.ColorTextLight)).
		Background(lipgloss.Color(ui.ColorPrimary)).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color(ui.ColorText))
	t.SetStyles(s)

	m.table = t
	m.adjustLayout()
	m.updateTable()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("zap - File Registry")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle editor finished messages globally
	if statusStr, ok := editor.HandleEditorFinished(msg); ok {
		return m, showStatus(statusStr)
	}

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
		switch m.mode {
		case ModeHelp:
			return m.updateHelp(msg)
		case ModeEdit, ModeAdd:
			return m.updateEdit(msg)
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeConfirmDelete:
			return m.updateDeleteConfirm(msg)
		default:
			return m.updateNormal(msg)
		}
	}

	// Let table handle mouse events when in normal mode
	if m.mode == ModeNormal {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key exits help
	m.mode = ModeNormal
	return m, nil
}

func (m model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.deleteIndex >= 0 && m.deleteIndex < len(m.configs) {
			configName := m.configs[m.deleteIndex].Name
			m.configs = append(m.configs[:m.deleteIndex], m.configs[m.deleteIndex+1:]...)
			if err := m.storage.Save(m.configs); err != nil {
				m.mode = ModeNormal
				return m, showStatus(fmt.Sprintf("❌ Failed to save: %v", err))
			}
			m.cacheValid = false
			m.updateTable()
			m.mode = ModeNormal
			m.deleteIndex = -1
			return m, showStatus(fmt.Sprintf("🗑️  Deleted %s", configName))
		}
		m.mode = ModeNormal
		m.deleteIndex = -1
		return m, nil
	case "n", "N", "esc":
		m.mode = ModeNormal
		m.deleteIndex = -1
		return m, showStatus("❌ Deletion cancelled")
	}
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.searchQuery = ""
		m.fuzzyMode = false
		m.searchInput.SetValue("")
		m.searchInput.Blur()
		m.cacheValid = false
		m.updateTable()
		return m, showStatus("🔍 Search cleared")
	case "enter":
		m.mode = ModeNormal
		m.searchQuery = m.searchInput.Value()
		m.searchInput.Blur()
		m.cacheValid = false
		m.updateTable()
		if m.searchQuery != "" {
			matchCount := m.getFilteredConfigsCount()
			return m, showStatus(fmt.Sprintf("🔍 Found %d matches", matchCount))
		}
		return m, nil
	}

	m.searchInput, cmd = m.searchInput.Update(msg)
	// Live update search results
	m.searchQuery = m.searchInput.Value()
	m.cacheValid = false
	m.updateTable()
	return m, cmd
}

func (m model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.cancelEdit()
		return m, nil
	case "enter":
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("❌ Failed to save: %v", err))
		}
		m.cancelEdit()
		if m.mode == ModeAdd {
			return m, showStatus("✅ File added")
		}
		return m, showStatus("✅ File updated")
	case "tab":
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("❌ Failed to save: %v", err))
		}
		m.editCol = (m.editCol + 1) % 5
		m.loadEditField()
		return m, nil
	case "shift+tab":
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("❌ Failed to save: %v", err))
		}
		m.editCol = (m.editCol - 1 + 5) % 5
		m.loadEditField()
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "?":
		m.mode = ModeHelp
		return m, nil

	case "/":
		m.mode = ModeSearch
		m.fuzzyMode = false
		m.searchInput.Focus()
		m.searchInput.SetValue(m.searchQuery)
		return m, showStatus("🔍 Search mode (press Enter to apply, Esc to cancel)")

	case "ctrl+f":
		m.mode = ModeSearch
		m.fuzzyMode = true
		m.searchInput.Focus()
		m.searchInput.Placeholder = "Fuzzy find..."
		m.searchInput.SetValue(m.searchQuery)
		return m, showStatus("🎯 Fuzzy find mode")

	case "s":
		m.sortByRecent = !m.sortByRecent
		m.cacheValid = false
		m.updateTable()
		if m.sortByRecent {
			return m, showStatus("📊 Sorted by recently opened")
		}
		return m, showStatus("📊 Sorted by project")

	case "e":
		return m, m.startEdit()

	case "n", "a":
		return m, m.addNewConfig()

	case "d", "delete":
		if len(m.configs) > 0 {
			displayIndex := m.table.Cursor()
			originalIndex := m.getOriginalIndexByDisplayIndex(displayIndex)
			if originalIndex == -1 {
				return m, nil
			}
			m.mode = ModeConfirmDelete
			m.deleteIndex = originalIndex
			return m, showStatus(fmt.Sprintf("❓ Delete '%s'? (y/n)", m.configs[originalIndex].Name))
		}
		return m, nil

	case " ", "enter":
		if len(m.configs) > 0 {
			displayIndex := m.table.Cursor()
			config := m.getConfigByDisplayIndex(displayIndex)
			if config != nil {
				// Update last opened time
				for i := range m.configs {
					if m.configs[i].Equals(config) {
						m.configs[i].LastOpened = time.Now()
						m.storage.Save(m.configs)
						m.cacheValid = false
						break
					}
				}
				return m, editor.OpenConfig(*config)
			}
		}
		return m, nil

	case "r":
		configs, err := m.storage.Load()
		if err != nil {
			return m, showStatus(fmt.Sprintf("❌ Failed to reload: %v", err))
		}
		m.configs = configs
		m.cacheValid = false
		m.updateTable()
		return m, showStatus("🔄 Refreshed")

	case "left", "h":
		if m.scrollOffset > 0 {
			m.scrollOffset--
			m.adjustLayout()
			m.updateTable()
		}
		return m, nil

	case "right", "l":
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

	case "k", "up":
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case "j", "down":
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case "g", "home":
		m.table.SetCursor(0)
		return m, nil

	case "G", "end":
		m.table.SetCursor(len(m.table.Rows()) - 1)
		return m, nil

	case "pageup", "ctrl+u":
		cursor := m.table.Cursor()
		pageSize := m.table.Height()
		newCursor := cursor - pageSize
		if newCursor < 0 {
			newCursor = 0
		}
		m.table.SetCursor(newCursor)
		return m, nil

	case "pagedown", "ctrl+d":
		cursor := m.table.Cursor()
		pageSize := m.table.Height()
		newCursor := cursor + pageSize
		maxCursor := len(m.table.Rows()) - 1
		if newCursor > maxCursor {
			newCursor = maxCursor
		}
		m.table.SetCursor(newCursor)
		return m, nil

	default:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
}

func (m *model) startEdit() tea.Cmd {
	if len(m.configs) == 0 {
		return showStatus("❌ No files to edit")
	}

	displayIndex := m.table.Cursor()
	m.editRow = m.getOriginalIndexByDisplayIndex(displayIndex)
	if m.editRow == -1 {
		return showStatus("❌ Invalid selection")
	}

	m.mode = ModeEdit
	m.editCol = 0
	m.loadEditField()
	m.textInput.Focus()
	return nil
}

func (m *model) addNewConfig() tea.Cmd {
	newConfig := models.ConfigEntry{
		Name:        "New File",
		Path:        "~/path/to/file",
		Type:        "txt",
		Project:     "",
		Description: "File description",
	}

	m.configs = append(m.configs, newConfig)
	m.cacheValid = false
	m.mode = ModeAdd
	m.editRow = len(m.configs) - 1
	m.editCol = 0
	m.loadEditField()
	m.textInput.Focus()
	m.updateTable()

	// Move cursor to new entry
	displayIndex := m.findConfigDisplayIndex(newConfig)
	if displayIndex != -1 {
		m.table.SetCursor(displayIndex)
	}

	return showStatus("➕ Adding new file (Tab to next field, Enter to save)")
}

func (m *model) loadEditField() {
	if m.editRow < 0 || m.editRow >= len(m.configs) {
		return
	}

	config := m.configs[m.editRow]
	var value string
	switch m.editCol {
	case 0:
		value = config.Name
	case 1:
		value = config.Project
	case 2:
		value = config.Type
	case 3:
		value = config.Path
	case 4:
		value = config.Description
	}

	m.textInput.SetValue(value)
	m.textInput.SetCursor(len(value))
}

func (m *model) saveEdit() error {
	if m.editRow < 0 || m.editRow >= len(m.configs) {
		return fmt.Errorf("invalid edit row")
	}

	value := strings.TrimSpace(m.textInput.Value())

	switch m.editCol {
	case 0: // Name
		if value == "" {
			return fmt.Errorf("name cannot be empty")
		}
		m.configs[m.editRow].Name = value
	case 1: // Project
		m.configs[m.editRow].Project = value
	case 2: // Type
		m.configs[m.editRow].Type = value
	case 3: // Path
		if value == "" {
			return fmt.Errorf("path cannot be empty")
		}
		expandedPath := editor.ExpandPath(value)

		// Check for duplicates (only when adding or changing path)
		if m.mode == ModeAdd || m.configs[m.editRow].Path != expandedPath {
			if dup := storage.FindDuplicates(m.configs, expandedPath); dup != nil && !dup.Equals(&m.configs[m.editRow]) {
				return fmt.Errorf("file already registered as '%s'", dup.Name)
			}
		}

		m.configs[m.editRow].Path = expandedPath

		// Auto-detect file type if not set or is default
		if m.configs[m.editRow].Type == "" || m.configs[m.editRow].Type == "txt" {
			m.configs[m.editRow].Type = models.DetectFileType(expandedPath)
		}
	case 4: // Description
		m.configs[m.editRow].Description = value
	}

	if err := m.storage.Save(m.configs); err != nil {
		return err
	}

	m.cacheValid = false
	m.updateTable()
	return nil
}

func (m *model) cancelEdit() {
	// If we were adding and canceled, remove the entry
	if m.mode == ModeAdd && m.editRow >= 0 && m.editRow < len(m.configs) {
		if m.configs[m.editRow].Path == "~/path/to/file" {
			m.configs = append(m.configs[:m.editRow], m.configs[m.editRow+1:]...)
			m.cacheValid = false
		}
	}

	m.mode = ModeNormal
	m.editRow = -1
	m.editCol = -1
	m.textInput.Blur()
	m.textInput.SetValue("")
	m.updateTable()
}

func (m *model) getSortedConfigs() []models.ConfigEntry {
	if m.cacheValid && m.sortedCache != nil {
		return m.sortedCache
	}

	var sorted []models.ConfigEntry
	if m.sortByRecent {
		sorted = storage.SortByRecentlyOpened(m.configs)
	} else {
		sorted = storage.SortConfigs(m.configs)
	}

	m.sortedCache = sorted
	m.cacheValid = true
	return sorted
}

func (m *model) getFilteredConfigs() []models.ConfigEntry {
	sorted := m.getSortedConfigs()

	if m.searchQuery == "" {
		return sorted
	}

	query := strings.ToLower(m.searchQuery)
	var filtered []models.ConfigEntry

	for _, config := range sorted {
		if m.matchesSearch(config, query) {
			filtered = append(filtered, config)
		}
	}

	return filtered
}

func (m *model) getFilteredConfigsCount() int {
	return len(m.getFilteredConfigs())
}

func (m *model) matchesSearch(config models.ConfigEntry, query string) bool {
	if m.fuzzyMode {
		return fuzzyMatch(query, strings.ToLower(config.Name)) ||
			fuzzyMatch(query, strings.ToLower(config.Project)) ||
			fuzzyMatch(query, strings.ToLower(config.Path)) ||
			fuzzyMatch(query, strings.ToLower(config.Description))
	}

	// Normal substring search
	return strings.Contains(strings.ToLower(config.Name), query) ||
		strings.Contains(strings.ToLower(config.Project), query) ||
		strings.Contains(strings.ToLower(config.Type), query) ||
		strings.Contains(strings.ToLower(config.Path), query) ||
		strings.Contains(strings.ToLower(config.Description), query)
}

func fuzzyMatch(pattern, text string) bool {
	if pattern == "" {
		return true
	}
	if text == "" {
		return false
	}

	patternIdx := 0
	textIdx := 0

	for textIdx < len(text) && patternIdx < len(pattern) {
		if text[textIdx] == pattern[patternIdx] {
			patternIdx++
		}
		textIdx++
	}

	return patternIdx == len(pattern)
}

func (m *model) updateTable() {
	filteredConfigs := m.getFilteredConfigs()
	visibleColumns := m.table.Columns()

	var rows []table.Row
	m.configIndices = []int{}

	var lastProject string
	configIndex := 0

	for _, config := range filteredConfigs {
		displayProject := config.Project
		if displayProject == "" {
			displayProject = "General"
		}

		// Add project header (only in non-search mode or when not sorting by recent)
		if !m.sortByRecent && displayProject != lastProject {
			projectHeader := fmt.Sprintf("📂 %s", displayProject)
			headerRow := make(table.Row, len(visibleColumns))
			headerRow[0] = projectHeader
			for i := 1; i < len(headerRow); i++ {
				headerRow[i] = ""
			}
			rows = append(rows, headerRow)
			m.configIndices = append(m.configIndices, -1)
			lastProject = displayProject
		}

		// Build row with status indicators
		name := config.Name
		if !editor.FileExists(config.Path) {
			name = "❌ " + name
		} else if !config.LastOpened.IsZero() {
			name = "✓ " + name
		}

		fullRowData := []string{name, displayProject, config.Type, config.Path, config.Description}

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

	availableWidth := m.width - 6

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

	if visibleCols == 0 {
		visibleCols = 1
		firstCol := m.allColumns[0]
		firstCol.Width = availableWidth
		m.allColumns[0] = firstCol
	}

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

	var visibleColumns []table.Column
	for i := startCol; i < endCol && i < len(m.allColumns); i++ {
		visibleColumns = append(visibleColumns, m.allColumns[i])
	}

	if len(visibleColumns) > 0 {
		usedWidth := 0
		for _, col := range visibleColumns {
			usedWidth += col.Width
		}
		if extraWidth := availableWidth - usedWidth; extraWidth > 0 {
			visibleColumns[len(visibleColumns)-1].Width += extraWidth
		}
	}

	m.table.SetColumns(visibleColumns)
	m.table.SetHeight(tableHeight)
	m.maxCols = len(m.allColumns)
}

func (m *model) getConfigByDisplayIndex(displayIndex int) *models.ConfigEntry {
	if displayIndex < 0 || displayIndex >= len(m.configIndices) {
		return nil
	}

	configIndex := m.configIndices[displayIndex]
	if configIndex == -1 {
		return nil
	}

	filteredConfigs := m.getFilteredConfigs()
	if configIndex >= len(filteredConfigs) {
		return nil
	}

	sortedConfig := filteredConfigs[configIndex]
	for i := range m.configs {
		if m.configs[i].Equals(&sortedConfig) {
			return &m.configs[i]
		}
	}
	return nil
}

func (m *model) findConfigDisplayIndex(targetConfig models.ConfigEntry) int {
	for i, configIndex := range m.configIndices {
		if configIndex == -1 {
			continue
		}

		filteredConfigs := m.getFilteredConfigs()
		if configIndex < len(filteredConfigs) {
			config := filteredConfigs[configIndex]
			if config.Equals(&targetConfig) {
				return i
			}
		}
	}
	return -1
}

func (m *model) getOriginalIndexByDisplayIndex(displayIndex int) int {
	if displayIndex < 0 || displayIndex >= len(m.configIndices) {
		return -1
	}

	configIndex := m.configIndices[displayIndex]
	if configIndex == -1 {
		return -1
	}

	filteredConfigs := m.getFilteredConfigs()
	if configIndex >= len(filteredConfigs) {
		return -1
	}

	sortedConfig := filteredConfigs[configIndex]
	for i := range m.configs {
		if m.configs[i].Equals(&sortedConfig) {
			return i
		}
	}
	return -1
}

func (m model) View() string {
	// Help mode
	if m.mode == ModeHelp {
		return ui.HelpScreen(m.width, m.height)
	}

	// Normal view
	header := ui.TitleStyle.Render("⚡ zap - File Registry")

	// Empty state
	if len(m.configs) == 0 {
		content := ui.MutedStyle.
			MarginTop(1).
			MarginBottom(1).
			Render("📋 No files registered yet.\n\n💡 Press 'n' to add your first file!")

		footer := ui.InfoStyle.Render("Commands: ") +
			ui.WarningStyle.Render("n/a: add file") +
			ui.MutedStyle.Render(" • ") +
			ui.ErrorStyle.Render("q: quit") +
			ui.MutedStyle.Render(" • ") +
			ui.InfoStyle.Render("?: help")

		return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
	}

	// Status message
	var statusMessage string
	if m.statusMsg != "" && time.Now().Before(m.statusExpiry) {
		style := ui.GetStatusStyle(m.statusMsg)
		statusMessage = style.Render("Status: " + m.statusMsg)
	}

	// Footer based on mode
	var footer string
	switch m.mode {
	case ModeEdit, ModeAdd:
		colNames := []string{"Name", "Project", "Type", "Path", "Description"}
		colName := colNames[m.editCol]

		typeHelp := ""
		if colName == "Type" {
			typeHelp = " (json, yaml, toml, etc.)"
		} else if colName == "Path" {
			typeHelp = " (~/ for home)"
		}

		prefix := "✏️  Editing"
		if m.mode == ModeAdd {
			prefix = "➕ Adding"
		}

		footer = ui.EditStyle.Render(fmt.Sprintf("%s %s%s: %s", prefix, colName, typeHelp, m.textInput.View())) +
			ui.HelpStyle.Render("\nCommands: tab: next field • enter: save • esc: cancel")

	case ModeSearch:
		searchType := "Search"
		if m.fuzzyMode {
			searchType = "Fuzzy Find"
		}
		matchCount := m.getFilteredConfigsCount()
		footer = ui.SearchStyle.Render(fmt.Sprintf("🔍 %s: %s", searchType, m.searchInput.View())) +
			ui.HelpStyle.Render(fmt.Sprintf("\n%d matches • enter: apply • esc: cancel", matchCount))

	case ModeConfirmDelete:
		footer = ui.DeleteStyle.Render(fmt.Sprintf("🗑️  Delete '%s'? ", m.configs[m.deleteIndex].Name)) +
			ui.HelpStyle.Render("y: yes • n/esc: no")

	default:
		commands := []string{
			ui.InfoStyle.Render("↑↓/jk: nav"),
			ui.SuccessStyle.Render("space: open"),
			ui.WarningStyle.Render("e: edit"),
			ui.WarningStyle.Render("n: add"),
			ui.ErrorStyle.Render("d: del"),
			ui.InfoStyle.Render("/: search"),
			ui.InfoStyle.Render("s: sort"),
			ui.ErrorStyle.Render("q: quit"),
			ui.InfoStyle.Render("?: help"),
		}

		scrollHint := ""
		if m.maxCols > len(m.table.Columns()) {
			scrollHint = " • " + ui.InfoStyle.Render("←→: scroll")
		}

		searchHint := ""
		if m.searchQuery != "" {
			searchHint = " • " + ui.SearchStyle.Render(fmt.Sprintf("🔍 '%s'", m.searchQuery))
		}

		footer = ui.HelpStyle.Render("Commands: "+strings.Join(commands, " • ")) + scrollHint + searchHint
	}

	// Build final view
	var parts []string
	parts = append(parts, header)
	parts = append(parts, m.table.View())

	if statusMessage != "" {
		parts = append(parts, statusMessage)
	}

	parts = append(parts, footer)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
