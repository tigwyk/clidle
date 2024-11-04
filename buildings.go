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
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/soft-serve/pkg/ui/common"
)

const (
	padding  = 2
	maxWidth = 80
)

// BuildingsMsg is sent when the tab is ready
type BuildingsMsg *BuildingsModel

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

// BuildingsModel is the Buildings component page
type BuildingsModel struct {
	game      *Game
	common    common.Common
	spinner   spinner.Model
	list      list.Model
	progress  progress.Model
	buildings []Building
	isLoading bool
}

// Path implements common.TabComponent.
func (m *BuildingsModel) Path() string {
	return "/buildings"
}

// TabName returns the name of the tab.
func (m *BuildingsModel) TabName() string {
	return "Buildings"
}

// Tick returns a command that ticks the spinner.
func (m *BuildingsModel) Tick() tea.Cmd {
	var cmds []tea.Cmd
	if m.progress.Percent() == 1.0 {
		log.Debug("Buildings progress bar hit 100%")
		m.progress.SetPercent(0)
	}

	// Note that you can also use progress.Model.SetPercent to set the
	// percentage value explicitly, too.
	cmd := tickBuildings(m)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	//return tea.Batch(m.spinner.Tick, m.updateBuildingsCmd, cmd, buildingsTickCmd)
	cmds = append(cmds, m.spinner.Tick)
	cmds = append(cmds, m.updateBuildingsCmd)
	// cmds = append(cmds, tickCmd())
	return tea.Batch(cmds...)
}

func tickBuildings(m *BuildingsModel) tea.Cmd {
	cmd := m.progress.IncrPercent(0.25)
	return cmd
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
	log.Debug("Buildings Init", "BuildingsModel", m)
	return m.Tick()
}

// Update updates the buildings tab.
func (m *BuildingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.game != nil && m.game.dump != nil {
		m.game.dump.Debug("Buildings Update", "msg", msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedItem := m.list.SelectedItem().(BuildingItem)
			selectedItem.Building.Level++
			m.updateBuilding(selectedItem.Building)
		}
	case GameMsg:
		m.game = msg
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
	case tickMsg:
		return m, m.Tick()
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
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
	log.Debug("NewBuildingsModel", "items", items)
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
	log.Debug("Updating buildings")
	if m.buildings == nil {
		log.Errorf("missing buildings")
		return common.ErrorMsg(common.ErrMissingRepo)
	}
	m.isLoading = false
	return BuildingsMsg(m)
}
