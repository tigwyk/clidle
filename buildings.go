// buildings.go

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

// BuildingsMsg is a message sent when the readme is loaded.
type BuildingsMsg struct {
	Content string
	Path    string
}

// Building represents a building in the game.
type Building struct {
	Name  string
	Level int
	Cost  int
}

// BuildingItem is a wrapper for Building to implement list.Item interface.
type BuildingItem struct {
	Building Building
}

func (i BuildingItem) Title() string { return i.Building.Name }
func (i BuildingItem) Description() string {
	return fmt.Sprintf("Level: %d, Cost: %d", i.Building.Level, i.Building.Cost)
}
func (i BuildingItem) FilterValue() string { return i.Building.Name }

// BuildingsModel is the model for the buildings tab.
type BuildingsModel struct {
	game      gameInfo
	common    common.Common
	spinner   spinner.Model
	list      list.Model
	progress  progress.Model
	buildings []Building
	isLoading bool
}

// Path implements common.TabComponent.
func (m *BuildingsModel) Path() string {
	return ""
}

// TabName returns the name of the tab.
func (m *BuildingsModel) TabName() string {
	return "Buildings"
}

// Tick returns a command that ticks the spinner.
func (m *BuildingsModel) Tick() tea.Cmd {
	return m.spinner.Tick
}

// SetSize implements common.Component.
func (m *BuildingsModel) SetSize(width, height int) {
	m.common.SetSize(width, height)
}

// ShortHelp implements help.KeyMap.
func (m *BuildingsModel) ShortHelp() []key.Binding {
	b := []key.Binding{
		m.common.KeyMap.UpDown,
	}
	return b
}

// FullHelp implements the common.TabComponent interface.
func (b *BuildingsModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			b.common.KeyMap.Back,
			b.common.KeyMap.Help,
		},
	}
}

// Init initializes the buildings tab.
func (m *BuildingsModel) Init() tea.Cmd {
	m.isLoading = true
	return tea.Batch(m.spinner.Tick, m.updateBuildingsCmd)
}

// Update updates the buildings tab.
func (m *BuildingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedItem := m.list.SelectedItem().(BuildingItem)
			selectedItem.Building.Level++
			m.updateBuilding(selectedItem.Building)
		}
	case GameMsg:
		m.game = msg.Game
	case BuildingsMsg:
		m.isLoading = false
	case spinner.TickMsg:
		if m.isLoading && m.spinner.ID() == msg.ID {
			s, cmd := m.spinner.Update(msg)
			m.spinner = s
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the buildings tab.
func (m *BuildingsModel) View() string {
	if m.isLoading {
		return renderLoading(m.common, m.spinner)
	} else {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.spinner.View(),
			m.list.View(),
			m.progress.View(),
		)
	}
}

// updateBuilding updates the building in the list.
func (m *BuildingsModel) updateBuilding(building Building) {
	for i, b := range m.buildings {
		if b.Name == building.Name {
			m.buildings[i] = building
			break
		}
	}
	m.updateList()
}

// updateList updates the list with the current buildings.
func (m *BuildingsModel) updateList() {
	items := make([]list.Item, len(m.buildings))
	for i, b := range m.buildings {
		items[i] = BuildingItem{Building: b}
	}
	m.list.SetItems(items)
}

// NewBuildingsModel returns a new buildings tab model.
func NewBuildingsModel(c common.Common) *BuildingsModel {
	buildings := []Building{
		{Name: "Building 1", Level: 1, Cost: 100},
		{Name: "Building 2", Level: 1, Cost: 200},
		{Name: "Building 3", Level: 1, Cost: 300},
	}
	items := make([]list.Item, len(buildings))
	for i, b := range buildings {
		items[i] = BuildingItem{Building: b}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Buildings"
	return &BuildingsModel{
		common:    c,
		spinner:   spinner.New(),
		list:      l,
		progress:  progress.New(progress.WithDefaultGradient()),
		buildings: buildings,
		isLoading: true,
	}
}

// SpinnerID implements common.TabComponent.
func (m *BuildingsModel) SpinnerID() int {
	return m.spinner.ID()
}

// StatusBarValue implements statusbar.StatusBar.
func (m *BuildingsModel) StatusBarValue() string {
	return "Buildings and stuff"
}

// StatusBarInfo implements statusbar.StatusBar.
func (m *BuildingsModel) StatusBarInfo() string {
	return fmt.Sprintf("☰ %d%%", m.list.Index())
}

func (m *BuildingsModel) updateBuildingsCmd() tea.Msg {
	bm := BuildingsMsg{}
	if m.buildings == nil {
		return common.ErrorMsg(common.ErrMissingRepo)
	}
	bm.Content = "fake content"
	bm.Path = "fake/path"
	return bm
}

var docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
