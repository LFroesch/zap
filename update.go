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

// reposViewMsg is a no-op message used to trigger textarea.repositionView()
// after manual cursor navigation (CursorUp/CursorDown don't update the viewport).
type reposViewMsg struct{}

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
		if m.mode == ModeFileEdit {
			m.resizeFileEditArea()
		}
		m.refreshRightViewport()
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.mode {
		case ModeHelp:
			return m.updateHelp(msg)
		case ModeEdit, ModeAdd:
			return m.updateEdit(msg)
		case ModeFileEdit:
			return m.updateFileEdit(msg)
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

func (m model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	pageSize := m.helpPageSize()

	switch msg.String() {
	case "esc", "?", "q":
		m.mode = ModeNormal
		m.helpScroll = 0
	case "up", "w", "k", "pgup":
		if msg.String() == "pgup" {
			m.helpScroll -= pageSize
		} else {
			m.helpScroll--
		}
	case "down", "s", "j", "pgdn":
		if msg.String() == "pgdn" {
			m.helpScroll += pageSize
		} else {
			m.helpScroll++
		}
	case "g", "home":
		m.helpScroll = 0
	case "G", "end":
		m.helpScroll = m.maxHelpScroll()
	}

	if m.helpScroll < 0 {
		m.helpScroll = 0
	}
	maxScroll := m.maxHelpScroll()
	if m.helpScroll > maxScroll {
		m.helpScroll = maxScroll
	}

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
			m.refreshRightViewport()
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
		m.refreshRightViewport()
		return m, showStatus("Search cleared")
	case "enter":
		m.mode = ModeNormal
		m.searchQuery = m.searchInput.Value()
		m.searchInput.Blur()
		m.cacheValid = false
		m.buildDisplayList()
		m.refreshRightViewport()
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
	m.refreshRightViewport()
	return m, cmd
}

func (m model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.cancelEdit()
		return m, nil
	case "enter":
		wasAdding := m.mode == ModeAdd
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("Failed to save: %v", err))
		}
		m.cancelEdit()
		if wasAdding {
			return m, showStatus("File added")
		}
		return m, showStatus("File updated")
	case "tab":
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("Failed to save: %v", err))
		}
		m.editCol = (m.editCol + 1) % 4
		m.loadEditField()
		return m, nil
	case "shift+tab":
		if err := m.saveEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("Failed to save: %v", err))
		}
		m.editCol = (m.editCol - 1 + 4) % 4
		m.loadEditField()
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateFileEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.cancelFileEdit()
		return m, showStatus("Inline edit cancelled")
	case "ctrl+s":
		if err := m.saveFileEdit(); err != nil {
			return m, showStatus(fmt.Sprintf("Save failed: %v", err))
		}
		return m, showStatus("File saved")
	case "ctrl+d":
		// Delete current line, reposition cursor to the same line number.
		value := m.fileEditArea.Value()
		lineNum := m.fileEditArea.Line()
		lines := strings.Split(value, "\n")
		if lineNum < len(lines) {
			newLines := append(lines[:lineNum], lines[lineNum+1:]...)
			m.fileEditArea.SetValue(strings.Join(newLines, "\n"))
			// SetValue leaves cursor at end; move up to target line.
			target := lineNum
			if target >= len(newLines) {
				target = len(newLines) - 1
			}
			if target < 0 {
				target = 0
			}
			for i := 0; i < len(newLines)-1-target; i++ {
				m.fileEditArea.CursorUp()
			}
			m.fileEditArea.CursorStart()
			// CursorUp doesn't call repositionView; pass a no-op msg to trigger it.
			m.fileEditArea, _ = m.fileEditArea.Update(reposViewMsg{})
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.fileEditArea, cmd = m.fileEditArea.Update(msg)
	return m, cmd
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	prevCursor := m.cursor

	switch msg.String() {
	case "q":
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
		m.sortMode = (m.sortMode + 1) % 4
		m.cacheValid = false
		m.buildDisplayList()
		m.refreshRightViewport()
		sortNames := []string{"Project", "Recent", "Name", "Path"}
		return m, showStatus(fmt.Sprintf("Sorted by %s", sortNames[m.sortMode]))

	case "e":
		return m, m.startEdit()

	case "E":
		return m, m.startFileEdit()

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
						m.buildDisplayList()
						break
					}
				}
				m.refreshRightViewport()
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
		m.refreshRightViewport()
		return m, showStatus("Refreshed")

	case "k", "up":
		m.moveCursorUp()

	case "j", "down":
		m.moveCursorDown()

	case "g":
		m.cursor = 0
		m.ensureCursorInBounds()

	case "G":
		m.cursor = len(m.displayConfigs) - 1
		m.ensureCursorInBounds()

	case "ctrl+u":
		pageSize := (m.height - uiOverhead) / 2
		for i := 0; i < pageSize; i++ {
			m.moveCursorUp()
		}

	case "ctrl+d":
		pageSize := (m.height - uiOverhead) / 2
		for i := 0; i < pageSize; i++ {
			m.moveCursorDown()
		}

	case "J", "s":
		m.rightViewport.LineDown(3)
		return m, nil

	case "K", "w":
		m.rightViewport.LineUp(3)
		return m, nil

	case "pgdown":
		m.rightViewport.ViewDown()
		return m, nil

	case "pgup":
		m.rightViewport.ViewUp()
		return m, nil

	case "ctrl+home":
		m.rightViewport.GotoTop()
		return m, nil

	case "ctrl+end":
		m.rightViewport.GotoBottom()
		return m, nil
	}

	if m.cursor != prevCursor {
		m.refreshRightViewport()
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
