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

// Weapon represents a weapon in the game.
type Weapon struct {
	Name  string
	Value int
}

// WeaponItem is a wrapper for Weapon to implement list.Item interface.
type WeaponItem struct {
	Weapon Weapon
}

func (i WeaponItem) Title() string { return i.Weapon.Name }
func (i WeaponItem) Description() string {
	return fmt.Sprintf("Value: %d", i.Weapon.Value)
}
func (i WeaponItem) FilterValue() string { return i.Weapon.Name }

// WeaponsModel is the model for the weapons tab.
type WeaponsModel struct {
	common   common.Common
	spinner  spinner.Model
	list     list.Model
	progress progress.Model
	weapons  []Weapon
}

// Path implements common.TabComponent.
func (m *WeaponsModel) Path() string {
	return ""
}

// TabName returns the name of the tab.
func (m *WeaponsModel) TabName() string {
	return "Weapons"
}

// Tick returns a command that ticks the spinner.
func (m *WeaponsModel) Tick() tea.Cmd {
	return m.spinner.Tick
}

// SetSize implements common.Component.
func (m *WeaponsModel) SetSize(width, height int) {
	m.common.SetSize(width, height)
}

// ShortHelp implements help.KeyMap.
func (m *WeaponsModel) ShortHelp() []key.Binding {
	b := []key.Binding{
		m.common.KeyMap.UpDown,
	}
	return b
}

// FullHelp implements the common.TabComponent interface.
func (m *WeaponsModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.common.KeyMap.Back,
			m.common.KeyMap.Help,
		},
	}
}

// Init initializes the weapons tab.
func (m *WeaponsModel) Init() tea.Cmd {
	return tea.Batch(
		m.Tick(),
	)
}

// Update updates the weapons tab.
func (m *WeaponsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedItem := m.list.SelectedItem().(WeaponItem)
			selectedItem.Weapon.Value++
			m.updateWeapon(selectedItem.Weapon)
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

// View renders the weapons tab.
func (m *WeaponsModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.spinner.View(),
		m.list.View(),
		m.progress.View(),
	)
}

// updateWeapon updates the weapon in the list.
func (m *WeaponsModel) updateWeapon(weapon Weapon) {
	for i, w := range m.weapons {
		if w.Name == weapon.Name {
			m.weapons[i] = weapon
			break
		}
	}
	m.updateList()
}

// updateList updates the list with the current weapons.
func (m *WeaponsModel) updateList() {
	items := make([]list.Item, len(m.weapons))
	for i, w := range m.weapons {
		items[i] = WeaponItem{Weapon: w}
	}
	m.list.SetItems(items)
}

// NewWeaponsModel returns a new weapons tab model.
func NewWeaponsModel(c common.Common) *WeaponsModel {
	weapons := []Weapon{
		{Name: "Weapon 1", Value: 1000},
		{Name: "Weapon 2", Value: 2000},
		{Name: "Weapon 3", Value: 3000},
	}
	items := make([]list.Item, len(weapons))
	for i, w := range weapons {
		items[i] = WeaponItem{Weapon: w}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Weapons"
	return &WeaponsModel{
		common:   c,
		spinner:  spinner.New(),
		list:     l,
		progress: progress.New(progress.WithDefaultGradient()),
		weapons:  weapons,
	}
}

// SpinnerID implements common.TabComponent.
func (m *WeaponsModel) SpinnerID() int {
	return m.spinner.ID()
}

// StatusBarValue implements statusbar.StatusBar.
func (m *WeaponsModel) StatusBarValue() string {
	return "Strength: 1000"
}

// StatusBarInfo implements statusbar.StatusBar.
func (m *WeaponsModel) StatusBarInfo() string {
	return fmt.Sprintf("☰ %d%%", m.list.Index())
}
