package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LFroesch/zap/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

var terminalEditors = map[string]bool{
	"nvim": true, "vim": true, "vi": true, "nano": true, "emacs": true,
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

// editorFinishedMsg is sent when the editor exits
type editorFinishedMsg struct {
	err  error
	name string
}

// OpenConfig opens a config file in the specified editor
func OpenConfig(config models.ConfigEntry, editorCmd string) tea.Cmd {
	return OpenPath(config.Path, editorCmd, config.Name)
}

// OpenPath opens any path in the specified editor
func OpenPath(path, editorCmd, label string) tea.Cmd {
	return func() tea.Msg {
		expandedPath := ExpandPath(path)

		if !FileExists(path) {
			return editorFinishedMsg{
				err:  fmt.Errorf("path not found: %s", expandedPath),
				name: label,
			}
		}

		if _, err := exec.LookPath(editorCmd); err != nil {
			return editorFinishedMsg{
				err:  fmt.Errorf("editor '%s' not found in PATH", editorCmd),
				name: label,
			}
		}

		isTerminal := terminalEditors[editorCmd]

		if isTerminal {
			cmd := exec.Command(editorCmd, expandedPath)
			return tea.ExecProcess(cmd, func(err error) tea.Msg {
				return editorFinishedMsg{err: err, name: label}
			})
		}

		cmd := exec.Command(editorCmd, expandedPath)
		err := cmd.Start()
		return editorFinishedMsg{err: err, name: label}
	}
}

// HandleEditorFinished processes the editor finished message
func HandleEditorFinished(msg tea.Msg) (string, bool) {
	if m, ok := msg.(editorFinishedMsg); ok {
		if m.err != nil {
			return fmt.Sprintf("Failed to open %s: %v", m.name, m.err), true
		}
		return fmt.Sprintf("Opened %s in editor", m.name), true
	}
	return "", false
}
