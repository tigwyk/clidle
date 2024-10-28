package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/soft-serve/pkg/ui/common"
)

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
	common   common.Common
	spinner  spinner.Model
	list     list.Model
	progress progress.Model
	capitals []Capital
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
	return m.spinner.Tick
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
	return tea.Batch(
		m.Tick(),
	)
}

// Update updates the capital tab.
func (m *CapitalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedItem := m.list.SelectedItem().(CapitalItem)
			selectedItem.Capital.Value++
			m.updateCapital(selectedItem.Capital)
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
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
	l.Title = "Capitals"
	return &CapitalModel{
		common:   c,
		spinner:  spinner.New(),
		list:     l,
		progress: progress.New(progress.WithDefaultGradient()),
		capitals: capitals,
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
