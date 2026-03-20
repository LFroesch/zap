package main

import (
	"fmt"
	"strings"

	"github.com/LFroesch/zap/internal/editor"
	"github.com/LFroesch/zap/internal/models"
	"github.com/LFroesch/zap/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

type statusMsg struct {
	message string
}

func showStatus(msg string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{message: msg}
	}
}

func (m *model) startEdit() tea.Cmd {
	if len(m.configs) == 0 {
		return showStatus("❌ No files to edit")
	}

	displayIndex := m.cursor
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
	m.buildDisplayList()

	// Move cursor to new entry
	displayIndex := m.findConfigDisplayIndex(newConfig)
	if displayIndex != -1 {
		m.cursor = displayIndex
		m.ensureCursorInBounds()
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
		value = config.Path
	case 3:
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
	case 2: // Path
		if value == "" {
			return fmt.Errorf("path cannot be empty")
		}
		expandedPath := editor.ExpandPath(value)

		if m.mode == ModeAdd || m.configs[m.editRow].Path != expandedPath {
			if dup := storage.FindDuplicates(m.configs, expandedPath); dup != nil && !dup.Equals(&m.configs[m.editRow]) {
				return fmt.Errorf("file already registered as '%s'", dup.Name)
			}
		}

		m.configs[m.editRow].Path = expandedPath

		// Auto-detect file type
		if m.configs[m.editRow].Type == "" || m.configs[m.editRow].Type == "txt" {
			m.configs[m.editRow].Type = models.DetectFileType(expandedPath)
		}
	case 3: // Description
		m.configs[m.editRow].Description = value
	}

	if err := m.storage.Save(m.configs); err != nil {
		return err
	}

	m.cacheValid = false
	m.buildDisplayList()
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
	m.buildDisplayList()
}

func (m *model) getSortedConfigs() []models.ConfigEntry {
	if m.cacheValid && m.sortedCache != nil {
		return m.sortedCache
	}

	var sorted []models.ConfigEntry
	switch m.sortMode {
	case 0: // Project
		sorted = storage.SortConfigs(m.configs)
	case 1: // Recent
		sorted = storage.SortByRecentlyOpened(m.configs)
	case 2: // Name
		sorted = storage.SortByName(m.configs)
	case 3: // Path
		sorted = storage.SortByPath(m.configs)
	default:
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

// buildDisplayList creates a flattened list of display items (headers + configs)
func (m *model) buildDisplayList() {
	filteredConfigs := m.getFilteredConfigs()
	m.displayConfigs = []displayConfig{}

	var lastProject string

	for i, config := range filteredConfigs {
		displayProject := config.Project
		if displayProject == "" {
			displayProject = "General"
		}

		// Add project header only when sorting by project (mode 0)
		if m.sortMode == 0 && displayProject != lastProject {
			m.displayConfigs = append(m.displayConfigs, displayConfig{
				isHeader:    true,
				headerText:  fmt.Sprintf("📂 %s", displayProject),
				configIndex: -1,
			})
			lastProject = displayProject
		}

		// Add config entry
		configCopy := config
		m.displayConfigs = append(m.displayConfigs, displayConfig{
			isHeader:    false,
			config:      &configCopy,
			configIndex: m.findOriginalIndex(config),
		})

		// Store original index mapping
		_ = i
	}

	m.ensureCursorInBounds()
}

// findOriginalIndex finds the index of a config in the original m.configs array
func (m *model) findOriginalIndex(target models.ConfigEntry) int {
	for i := range m.configs {
		if m.configs[i].Equals(&target) {
			return i
		}
	}
	return -1
}

func (m *model) ensureCursorInBounds() {
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.displayConfigs) {
		m.cursor = len(m.displayConfigs) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// Skip headers when cursor lands on them
	if m.cursor < len(m.displayConfigs) && m.displayConfigs[m.cursor].isHeader {
		// Move to next non-header
		for m.cursor < len(m.displayConfigs) && m.displayConfigs[m.cursor].isHeader {
			m.cursor++
		}
		if m.cursor >= len(m.displayConfigs) {
			// Went past the end, go back
			m.cursor = len(m.displayConfigs) - 1
			for m.cursor >= 0 && m.displayConfigs[m.cursor].isHeader {
				m.cursor--
			}
		}
	}
}

func (m *model) getConfigByDisplayIndex(displayIndex int) *models.ConfigEntry {
	if displayIndex < 0 || displayIndex >= len(m.displayConfigs) {
		return nil
	}

	display := m.displayConfigs[displayIndex]
	if display.isHeader || display.configIndex == -1 {
		return nil
	}

	if display.configIndex >= 0 && display.configIndex < len(m.configs) {
		return &m.configs[display.configIndex]
	}

	return nil
}

func (m *model) findConfigDisplayIndex(targetConfig models.ConfigEntry) int {
	for i, display := range m.displayConfigs {
		if !display.isHeader && display.config != nil && display.config.Equals(&targetConfig) {
			return i
		}
	}
	return -1
}

func (m *model) getOriginalIndexByDisplayIndex(displayIndex int) int {
	if displayIndex < 0 || displayIndex >= len(m.displayConfigs) {
		return -1
	}

	display := m.displayConfigs[displayIndex]
	if display.isHeader {
		return -1
	}

	return display.configIndex
}

// moveCursorUp moves cursor up, skipping headers
func (m *model) moveCursorUp() {
	if m.cursor <= 0 {
		return
	}

	m.cursor--
	// Skip headers
	for m.cursor >= 0 && m.displayConfigs[m.cursor].isHeader {
		m.cursor--
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureCursorInBounds()
}

// moveCursorDown moves cursor down, skipping headers
func (m *model) moveCursorDown() {
	if m.cursor >= len(m.displayConfigs)-1 {
		return
	}

	m.cursor++
	// Skip headers
	for m.cursor < len(m.displayConfigs) && m.displayConfigs[m.cursor].isHeader {
		m.cursor++
	}
	if m.cursor >= len(m.displayConfigs) {
		m.cursor = len(m.displayConfigs) - 1
	}
	m.ensureCursorInBounds()
}
