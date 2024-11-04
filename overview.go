// overview.go

package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/soft-serve/pkg/ui/common"
)

// OverviewModel represents the Overview tab.
type OverviewModel struct {
	common    common.Common
	progress  progress.Model
	spinner   spinner.Model
	isLoading bool
	game      *Game
}

const (
	overviewTabName = "Overview"
)

// NewOverviewModel creates a new OverviewModel.
func NewOverviewModel(c common.Common) *OverviewModel {
	return &OverviewModel{
		common:    c,
		progress:  progress.New(progress.WithDefaultGradient()),
		spinner:   spinner.New(),
		isLoading: true,
	}
}

// Init initializes the OverviewModel.
func (m *OverviewModel) Init() tea.Cmd {
	m.isLoading = true
	return m.Tick()
}

// Tick returns a command that advances the progress bar.
func (m *OverviewModel) Tick() tea.Cmd {
	var cmds []tea.Cmd
	if m.progress.Percent() >= 1.0 {
		log.Debug("Overview progress bar hit 100%")
		m.progress.SetPercent(0)
		m.incrementMoney()
	}
	cmds = append(cmds, m.progress.IncrPercent(0.1))
	cmds = append(cmds, m.spinner.Tick)
	cmds = append(cmds, tickCmd())
	return tea.Batch(cmds...)
}

// Update updates the OverviewModel.
func (m *OverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key messages if needed.
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
	case tickMsg:
		return m, m.Tick()
	case progress.FrameMsg:
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.progress.Update(msg)
		m.progress = model.(progress.Model)
		cmds = append(cmds, cmd)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case GameMsg:
		m.game = msg
	}
	return m, tea.Batch(cmds...)
}

// View renders the OverviewModel.
func (m *OverviewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		// fmt.Sprintf("Money: %d", m.game.money),
		fmt.Sprintf("Money: %d", 100),
		m.progress.View(),
	)
}

// incrementMoney calculates and adds money based on buildings owned.
func (m *OverviewModel) incrementMoney() {
	total := 0
	for _, tab := range m.game.panes {
		switch model := tab.(type) {
		case *BuildingsModel:
			for _, building := range model.buildings {
				total += building.Level * building.Cost
			}
		}
	}
	m.game.money += total
}

// TabName returns the name of the tab.
func (m *OverviewModel) TabName() string {
	return overviewTabName
}

// SetSize sets the size of the OverviewModel.
func (m *OverviewModel) SetSize(width, height int) {
	m.common.SetSize(width, height)
	m.progress.Width = width - padding*2 - 4
	if m.progress.Width > maxWidth {
		m.progress.Width = maxWidth
	}
}

// ShortHelp returns keybindings for the help menu.
func (m *OverviewModel) ShortHelp() []key.Binding {
	return nil
}

// FullHelp returns detailed keybindings for the help menu.
func (m *OverviewModel) FullHelp() [][]key.Binding {
	return nil
}

// Path returns the current path.
func (m *OverviewModel) Path() string {
	return "/" + overviewTabName
}

// SpinnerID returns the spinner ID.
func (m *OverviewModel) SpinnerID() int {
	return m.spinner.ID()
}

// StatusBarValue returns status bar value.
func (m *OverviewModel) StatusBarValue() string {
	return fmt.Sprintf("Money: %d", m.game.money)
}

// StatusBarInfo returns status bar info.
func (m *OverviewModel) StatusBarInfo() string {
	return "Overview Tab"
}
