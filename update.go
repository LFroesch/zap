package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LFroesch/zap/internal/editor"

	tea "github.com/charmbracelet/bubbletea"
)

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

	return m, nil
}

func (m model) updateHelp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
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
				return m, showStatus(fmt.Sprintf("Failed to save: %v", err))
			}
			m.cacheValid = false
			m.buildDisplayList()
			m.mode = ModeNormal
			m.deleteIndex = -1
			return m, showStatus(fmt.Sprintf("Deleted %s", configName))
		}
		m.mode = ModeNormal
		m.deleteIndex = -1
		return m, nil
	case "n", "N", "esc":
		m.mode = ModeNormal
		m.deleteIndex = -1
		return m, showStatus("Deletion cancelled")
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
		m.buildDisplayList()
		return m, showStatus("Search cleared")
	case "enter":
		m.mode = ModeNormal
		m.searchQuery = m.searchInput.Value()
		m.searchInput.Blur()
		m.cacheValid = false
		m.buildDisplayList()
		if m.searchQuery != "" {
			matchCount := m.getFilteredConfigsCount()
			return m, showStatus(fmt.Sprintf("Found %d matches", matchCount))
		}
		return m, nil
	}

	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = m.searchInput.Value()
	m.cacheValid = false
	m.buildDisplayList()
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
			return m, showStatus(fmt.Sprintf("Failed to save: %v", err))
		}
		m.cancelEdit()
		if m.mode == ModeAdd {
			return m, showStatus("File added")
		}
		return m, showStatus("File updated")
	case "tab":
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("Failed to save: %v", err))
		}
		m.editCol = (m.editCol + 1) % 5
		m.loadEditField()
		return m, nil
	case "shift+tab":
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("Failed to save: %v", err))
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
		return m, nil

	case "S":
		m.sortMode = (m.sortMode + 1) % 5
		m.cacheValid = false
		m.buildDisplayList()
		sortNames := []string{"Project", "Recent", "Name", "Type", "Path"}
		return m, showStatus(fmt.Sprintf("Sorted by %s", sortNames[m.sortMode]))

	case "e":
		return m, m.startEdit()

	case "N":
		return m, m.addNewConfig()

	case "D":
		if len(m.configs) > 0 {
			displayIndex := m.cursor
			originalIndex := m.getOriginalIndexByDisplayIndex(displayIndex)
			if originalIndex == -1 {
				return m, nil
			}
			m.mode = ModeConfirmDelete
			m.deleteIndex = originalIndex
			return m, showStatus(fmt.Sprintf("Delete '%s'? (y/n)", m.configs[originalIndex].Name))
		}
		return m, nil

	case "enter", "o":
		if len(m.configs) > 0 {
			displayIndex := m.cursor
			config := m.getConfigByDisplayIndex(displayIndex)
			if config != nil {
				for i := range m.configs {
					if m.configs[i].Equals(config) {
						m.configs[i].LastOpened = time.Now()
						m.storage.Save(m.configs)
						m.cacheValid = false
						break
					}
				}
				return m, editor.OpenConfig(*config, m.editor)
			}
		}
		return m, nil

	case "O":
		if len(m.configs) > 0 {
			config := m.getConfigByDisplayIndex(m.cursor)
			if config != nil {
				dir := filepath.Dir(editor.ExpandPath(config.Path))
				return m, editor.OpenPath(dir, m.editor, filepath.Base(dir))
			}
		}
		return m, nil

	case "y":
		if len(m.configs) > 0 {
			config := m.getConfigByDisplayIndex(m.cursor)
			if config != nil {
				path := editor.ExpandPath(config.Path)
				if err := copyToClipboard(path); err != nil {
					return m, showStatus(fmt.Sprintf("Clipboard error: %v", err))
				}
				return m, showStatus(fmt.Sprintf("Copied: %s", path))
			}
		}
		return m, nil

	case ",":
		configPath := m.storage.GetFilePath()
		return m, editor.OpenPath(configPath, m.editor, "zap config")

	case "r":
		configs, err := m.storage.Load()
		if err != nil {
			return m, showStatus(fmt.Sprintf("Failed to reload: %v", err))
		}
		m.configs = configs
		m.editor = m.storage.GetEditor()
		m.cacheValid = false
		m.buildDisplayList()
		return m, showStatus("Refreshed")

	case "k", "up":
		m.moveCursorUp()
		return m, nil

	case "j", "down":
		m.moveCursorDown()
		return m, nil

	case "g":
		m.cursor = 0
		m.ensureCursorInBounds()
		return m, nil

	case "G":
		m.cursor = len(m.displayConfigs) - 1
		m.ensureCursorInBounds()
		return m, nil

	case "ctrl+u":
		pageSize := (m.height - uiOverhead) / 2
		for i := 0; i < pageSize; i++ {
			m.moveCursorUp()
		}
		return m, nil

	case "ctrl+d":
		pageSize := (m.height - uiOverhead) / 2
		for i := 0; i < pageSize; i++ {
			m.moveCursorDown()
		}
		return m, nil
	}

	return m, nil
}

func copyToClipboard(text string) error {
	// Try clipboard commands in order of preference
	cmds := []struct {
		name string
		args []string
	}{
		{"wl-copy", nil},
		{"xclip", []string{"-selection", "clipboard"}},
		{"xsel", []string{"--clipboard", "--input"}},
		{"clip.exe", nil},
	}

	for _, c := range cmds {
		path, err := exec.LookPath(c.name)
		if err != nil {
			continue
		}
		cmd := exec.Command(path, c.args...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no clipboard tool found (tried wl-copy, xclip, xsel, clip.exe)")
}
