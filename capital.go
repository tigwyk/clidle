package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/soft-serve/pkg/ui/common"
)

// BuildingsMsg is a message sent when the readme is loaded.
type CapitalMsg *CapitalModel

// Capital represents a capital in the game.
type Capital struct {
	Name  string
	Value int
}

// CapitalItem is a wrapper for Capital to implement list.Item interface.
type CapitalItem struct {
	Capital Capital
}

func (i CapitalItem) Title() string { return i.Capital.Name }
func (i CapitalItem) Description() string {
	return fmt.Sprintf("Value: %d", i.Capital.Value)
}
func (i CapitalItem) FilterValue() string { return i.Capital.Name }

// CapitalModel is the model for the capital tab.
type CapitalModel struct {
	game      *Game
	common    common.Common
	spinner   spinner.Model
	list      list.Model
	progress  progress.Model
	capitals  []Capital
	isLoading bool
}

// Path implements common.TabComponent.
func (m *CapitalModel) Path() string {
	return ""
}

// TabName returns the name of the tab.
func (m *CapitalModel) TabName() string {
	return "Capital"
}

// Tick returns a command that ticks the spinner.
func (m *CapitalModel) Tick() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.updateCapitalsCmd)
}

// SetSize implements common.Component.
func (m *CapitalModel) SetSize(width, height int) {
	m.common.SetSize(width, height)
}

// ShortHelp implements help.KeyMap.
func (m *CapitalModel) ShortHelp() []key.Binding {
	b := []key.Binding{
		m.common.KeyMap.UpDown,
	}
	return b
}

// FullHelp implements the common.TabComponent interface.
func (m *CapitalModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.common.KeyMap.Back,
			m.common.KeyMap.Help,
		},
	}
}

// Init initializes the capital tab.
func (m *CapitalModel) Init() tea.Cmd {
	m.isLoading = true
	return m.Tick()
}

// Update updates the capital tab.
func (m *CapitalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.game != nil && m.game.dump != nil {
		m.game.dump.Debug("Capital Update", "msg", msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedItem := m.list.SelectedItem().(CapitalItem)
			selectedItem.Capital.Value++
			m.updateCapital(selectedItem.Capital)
		}
	case CapitalMsg:
		m.isLoading = false
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the capital tab.
func (m *CapitalModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.spinner.View(),
		m.list.View(),
		m.progress.View(),
	)
}

// updateCapital updates the capital in the list.
func (m *CapitalModel) updateCapital(capital Capital) {
	for i, c := range m.capitals {
		if c.Name == capital.Name {
			m.capitals[i] = capital
			break
		}
	}
	m.updateList()
}

// updateList updates the list with the current capitals.
func (m *CapitalModel) updateList() {
	items := make([]list.Item, len(m.capitals))
	for i, c := range m.capitals {
		items[i] = CapitalItem{Capital: c}
	}
	m.list.SetItems(items)
}

// NewCapitalModel returns a new capital tab model.
func NewCapitalModel(c common.Common) *CapitalModel {
	capitals := []Capital{
		{Name: "Capital 1", Value: 1000},
		{Name: "Capital 2", Value: 2000},
		{Name: "Capital 3", Value: 3000},
	}
	items := make([]list.Item, len(capitals))
	for i, c := range capitals {
		items[i] = CapitalItem{Capital: c}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	log.Debug("NewCapitalModel", "items", items)
	l.Title = "Capital"
	return &CapitalModel{
		common:    c,
		spinner:   spinner.New(),
		list:      l,
		progress:  progress.New(progress.WithDefaultGradient()),
		capitals:  capitals,
		isLoading: true,
	}
}

// SpinnerID implements common.TabComponent.
func (m *CapitalModel) SpinnerID() int {
	return m.spinner.ID()
}

// StatusBarValue implements statusbar.StatusBar.
func (m *CapitalModel) StatusBarValue() string {
	return "Money: $1000"
}

// StatusBarInfo implements statusbar.StatusBar.
func (m *CapitalModel) StatusBarInfo() string {
	return fmt.Sprintf("☰ %d%%", m.list.Index())
}

func (m *CapitalModel) updateCapitalsCmd() tea.Msg {
	log.Debug("Updating Capitals")
	if m.capitals == nil {
		log.Errorf("missing capitals")
		return common.ErrorMsg(common.ErrMissingRepo)
	}
	m.isLoading = false
	return CapitalMsg(m)
}
