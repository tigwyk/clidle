package main

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type item struct {
	Name  string
	Level int
	Cost  int
}

func (i item) FilterValue() string { return i.Name }

type TabModel struct {
	moneyProgress progress.Model
	Items         list.Model
	ItemTable     table.Model
	selectedRow   int
	originalRows  []table.Row
}

func (m TabModel) Init() tea.Cmd {
	return nil
}

func (m TabModel) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Items.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if m.selectedRow > 0 {
				m.selectedRow--
			}
		case "down":
			if m.selectedRow < len(m.ItemTable.Rows())-1 {
				m.selectedRow++
			}
		case "enter":
			// Handle item selection
			selectedItem := m.ItemTable.Rows()[m.selectedRow]
			fmt.Printf("Selected item: %v\n", selectedItem)
		}
	case tickMsg:
		return m, m.moneyProgress.IncrPercent(0.01)
	}
	m.ItemTable, cmd = m.ItemTable.Update(msg)
	return m, cmd
}

func (m TabModel) View() string {
	// Reset the rows to their original state
	rows := make([]table.Row, len(m.originalRows))
	copy(rows, m.originalRows)

	// Highlight the selected row
	if m.selectedRow >= 0 && m.selectedRow < len(rows) {
		rows[m.selectedRow][0] = fmt.Sprintf("> %s", rows[m.selectedRow][0])
	}

	m.ItemTable.SetRows(rows)
	return fmt.Sprintf("Primary Tab\n\n%s", m.ItemTable.View())
}

func newTabModel() TabModel {
	// Make initial list of items
	initialItems := []list.Item{
		item{Name: "Building 1", Level: 1, Cost: 50},
		item{Name: "Building 2", Level: 1, Cost: 100},
		item{Name: "Building 3", Level: 1, Cost: 200},
	}

	// Use reflection to get the field names of the item struct
	var columns []table.Column
	itemType := reflect.TypeOf(initialItems[0])
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		columns = append(columns, table.Column{Title: field.Name, Width: 20})
	}

	// Create rows dynamically based on the values of the fields
	var rows []table.Row
	for _, i := range initialItems {
		it := i.(item)
		row := table.Row{}
		itemValue := reflect.ValueOf(it)
		for j := 0; j < itemValue.NumField(); j++ {
			fieldValue := itemValue.Field(j)
			row = append(row, fmt.Sprintf("%v", fieldValue.Interface()))
		}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(7),
	)

	return TabModel{
		Items:         list.New(initialItems, list.NewDefaultDelegate(), 0, 0),
		moneyProgress: progress.New(progress.WithDefaultGradient()),
		ItemTable:     t,
		selectedRow:   0,
		originalRows:  rows,
	}
}
