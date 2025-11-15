package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"configly/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

// Editor represents an available editor
type Editor struct {
	Name       string
	Command    string
	IsTerminal bool
}

// AvailableEditors lists editors in order of preference
var AvailableEditors = []Editor{
	{Name: "VS Code", Command: "code", IsTerminal: false},
	{Name: "Neovim", Command: "nvim", IsTerminal: true},
	{Name: "Vim", Command: "vim", IsTerminal: true},
	{Name: "Nano", Command: "nano", IsTerminal: true},
	{Name: "Vi", Command: "vi", IsTerminal: true},
}

// FindAvailableEditor finds the first available editor on the system
func FindAvailableEditor() *Editor {
	for _, editor := range AvailableEditors {
		if _, err := exec.LookPath(editor.Command); err == nil {
			return &editor
		}
	}
	return nil
}

// ExpandPath expands ~ to home directory
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// FileExists checks if a file exists and is accessible
func FileExists(path string) bool {
	expandedPath := ExpandPath(path)
	_, err := os.Stat(expandedPath)
	return err == nil
}

// FileStatus returns status information about a file
type FileStatus struct {
	Exists     bool
	IsReadable bool
	Size       int64
}

// GetFileStatus returns detailed status about a file
func GetFileStatus(path string) FileStatus {
	expandedPath := ExpandPath(path)
	status := FileStatus{}

	info, err := os.Stat(expandedPath)
	if err != nil {
		return status
	}

	status.Exists = true
	status.Size = info.Size()

	// Check if readable
	file, err := os.Open(expandedPath)
	if err == nil {
		status.IsReadable = true
		file.Close()
	}

	return status
}

// editorFinishedMsg is sent when the editor exits
type editorFinishedMsg struct {
	err  error
	name string
}

// OpenConfig opens a config file in the appropriate editor
func OpenConfig(config models.ConfigEntry) tea.Cmd {
	return func() tea.Msg {
		expandedPath := ExpandPath(config.Path)

		// Check if file exists
		if !FileExists(config.Path) {
			return editorFinishedMsg{
				err:  fmt.Errorf("file not found: %s", expandedPath),
				name: config.Name,
			}
		}

		// Find available editor
		editor := FindAvailableEditor()
		if editor == nil {
			return editorFinishedMsg{
				err:  fmt.Errorf("no suitable editor found (tried: code, nvim, vim, nano, vi)"),
				name: config.Name,
			}
		}

		// For terminal editors, we need to use tea.ExecProcess
		// For GUI editors, we can use regular exec
		if editor.IsTerminal {
			cmd := exec.Command(editor.Command, expandedPath)
			return tea.ExecProcess(cmd, func(err error) tea.Msg {
				return editorFinishedMsg{err: err, name: config.Name}
			})
		} else {
			// GUI editor - start in background
			cmd := exec.Command(editor.Command, expandedPath)
			err := cmd.Start()
			return editorFinishedMsg{err: err, name: config.Name}
		}
	}
}

// HandleEditorFinished processes the editor finished message
func HandleEditorFinished(msg tea.Msg) (string, bool) {
	if m, ok := msg.(editorFinishedMsg); ok {
		if m.err != nil {
			return fmt.Sprintf("❌ Failed to open %s: %v", m.name, m.err), true
		}
		return fmt.Sprintf("📝 Opened %s in editor", m.name), true
	}
	return "", false
}
