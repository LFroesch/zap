package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/LFroesch/zap/internal/storage"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

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
		editor:       store.GetEditor(),
		width:        100,
		height:       24,
		mode:         ModeNormal,
		cursor:       0,
		scrollOffset: 0,
		editRow:      -1,
		editCol:      -1,
		deleteIndex:  -1,
		cacheValid:   false,
		sortMode:     0, // Start with Project sort
	}

	// Initialize text inputs
	m.textInput = textinput.New()
	m.textInput.CharLimit = 300

	m.fileEditArea = textarea.New()
	m.fileEditArea.CharLimit = 0

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Type to search..."
	m.searchInput.CharLimit = 100

	m.rightViewport = viewport.New(40, 10)

	// Build initial display list
	m.buildDisplayList()
	m.refreshRightViewport()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("zap - File Registry")
}
