package main

import (
	"time"

	"github.com/LFroesch/zap/internal/models"
	"github.com/LFroesch/zap/internal/storage"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// ViewMode represents the different view states
type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeEdit
	ModeAdd
	ModeFileEdit
	ModeSearch
	ModeHelp
	ModeConfirmDelete
)

const (
	uiOverhead = 9 // Header (1) + status (1) + borders (4) + padding (3)
)

type model struct {
	configs []models.ConfigEntry
	storage *storage.Storage
	editor  string
	width   int
	height  int

	// Navigation
	cursor       int
	scrollOffset int

	// Mode management
	mode ViewMode

	// Help mode
	helpScroll int

	// Edit mode
	editRow       int
	editCol       int
	textInput     textinput.Model
	fileEditArea  textarea.Model
	fileEditPath  string
	fileEditLabel string

	// Search mode
	searchInput textinput.Model
	searchQuery string
	fuzzyMode   bool

	// Delete confirmation
	deleteIndex int

	// UI state
	statusMsg    string
	statusExpiry time.Time

	// Display data
	displayConfigs []displayConfig // Flattened list with headers
	rightViewport  viewport.Model

	// Performance
	sortedCache []models.ConfigEntry
	cacheValid  bool
	sortMode    int // 0=Project, 1=Recent, 2=Name, 3=Type, 4=Path
}

// displayConfig represents a row in the display (either a header or a config)
type displayConfig struct {
	isHeader    bool
	headerText  string
	config      *models.ConfigEntry
	configIndex int // Index in m.configs (-1 for headers)
}
