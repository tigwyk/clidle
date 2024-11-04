package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/soft-serve/pkg/ui/common"
	"github.com/charmbracelet/soft-serve/pkg/ui/components/selector"
	"github.com/charmbracelet/soft-serve/pkg/ui/components/statusbar"
	"github.com/charmbracelet/soft-serve/pkg/ui/components/tabs"
)

var docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)

type state int

// GoBackMsg is a message to go back to the previous view.
type GoBackMsg struct{}

// GameMsg is a message to update the game.
type GameMsg *Game

type tickMsg time.Time

// SwitchTabMsg is a message to switch tabs.
type SwitchTabMsg common.TabComponent

type gameInfo struct {
	Name        string
	Description string
	ProjectName string
}

const (
	loadingState state = iota
	readyState
)

type Game struct {
	saveFile   string
	game       gameInfo
	common     common.Common
	tabs       *tabs.Tabs
	activeTab  int
	spinner    spinner.Model
	statusbar  *statusbar.Model
	panes      []common.TabComponent
	state      state
	panesReady []bool
	dump       *log.Logger
	money      int
}

// New returns a new Game.
func newGame(c common.Common, comps ...common.TabComponent) *Game {
	sb := statusbar.New(c)
	ts := make([]string, 0)
	for _, c := range comps {
		ts = append(ts, c.TabName())
	}
	tb := tabs.New(c, ts)
	// Make sure the order matches the order of tab constants above.
	s := spinner.New(spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(c.Styles.Spinner))

	gi := gameInfo{
		Name:        "Lord of War",
		Description: "A game about money and guns",
		ProjectName: "lord-of-war",
	}
	g := &Game{
		common:     c,
		tabs:       tb,
		statusbar:  sb,
		panes:      comps,
		state:      loadingState,
		spinner:    s,
		game:       gi,
		panesReady: make([]bool, len(comps)),
	}
	return g
}

func (g *Game) getMargins() (int, int) {
	hh := lipgloss.Height(g.headerView())
	hm := g.common.Styles.Repo.Body.GetVerticalFrameSize() +
		hh +
		g.common.Styles.Repo.Header.GetVerticalFrameSize() +
		g.common.Styles.StatusBar.GetHeight()
	return 0, hm
}

// SetSize implements common.Component.
func (g *Game) SetSize(width, height int) {
	g.common.SetSize(width, height)
	_, hm := g.getMargins()
	g.tabs.SetSize(width, height-hm)
	g.statusbar.SetSize(width, height-hm)
	for _, p := range g.panes {
		p.SetSize(width, height-hm)
	}
}

// Path returns the current component path.
func (g *Game) Path() string {
	return g.panes[g.activeTab].Path()
}

func (g *Game) commonHelp() []key.Binding {
	b := make([]key.Binding, 0)
	back := g.common.KeyMap.Back
	back.SetHelp("esc", "back to menu")
	tab := g.common.KeyMap.Section
	tab.SetHelp("tab", "switch tab")
	b = append(b, back)
	b = append(b, tab)
	return b
}

func (g *Game) Tick() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tickCmd())
	return tea.Batch(cmds...)
}

// ShortHelp implements help.KeyMap.
func (g *Game) ShortHelp() []key.Binding {
	b := g.commonHelp()
	b = append(b, g.panes[g.activeTab].(help.KeyMap).ShortHelp()...)
	return b
}

// FullHelp implements help.KeyMap.
func (g *Game) FullHelp() [][]key.Binding {
	b := make([][]key.Binding, 0)
	b = append(b, g.commonHelp())
	b = append(b, g.panes[g.activeTab].(help.KeyMap).FullHelp()...)
	return b
}

// Init implements tea.View.
func (g *Game) Init() tea.Cmd {
	log.Debug("Initializing game")
	g.state = loadingState
	g.activeTab = 0
	return tea.Batch(
		g.tabs.Init(),
		g.statusbar.Init(),
		g.panes[g.activeTab].Init(),
		g.spinner.Tick,
	)
}

func (g *Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	if g.dump != nil {
		g.dump.Debug("Game Update", "msg", msg)
	}
	switch msg := msg.(type) {
	case GameMsg:
		log.Debug("Received GameMsg")
		cmds = append(cmds,
			g.updateModels(msg),
		)
	case tea.WindowSizeMsg:
		g.SetSize(msg.Width, msg.Height)
		cmds = append(cmds, g.updateModels(msg))
	case tabs.SelectTabMsg:
		g.activeTab = int(msg)
		t, cmd := g.tabs.Update(msg)
		g.tabs = t.(*tabs.Tabs)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case tabs.ActiveTabMsg:
		g.activeTab = int(msg)
	case tea.KeyMsg, tea.MouseMsg:
		t, cmd := g.tabs.Update(msg)
		g.tabs = t.(*tabs.Tabs)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, g.common.KeyMap.Back):
				cmds = append(cmds, goBackCmd)
				// case key.Matches(msg, g.common.KeyMap.Select):
				// 	g.dump = log.FromContext(g.common.Context()).WithPrefix("Dump")
				// case key.Matches(msg, g.common.KeyMap.NextPage):
				// 	cmds = append(cmds, g.debugGameMsg)
			}
		}
	case BuildingsMsg:
		log.Debug("Received BuildingsMsg")
		cmds = append(cmds, g.updateTabComponent(&BuildingsModel{}, msg))
		if g.state == loadingState && !msg.isLoading {
			g.state = readyState
		}
	case CapitalMsg:
		log.Debug("Received CapitalMsg")
		cmds = append(cmds, g.updateTabComponent(&CapitalModel{}, msg))
		if g.state == loadingState && !msg.isLoading {
			g.state = readyState
		}
	case WeaponsMsg:
		log.Debug("Received WeaponsMsg")
		cmds = append(cmds, g.updateTabComponent(&WeaponsModel{}, msg))
		if g.state == loadingState && !msg.isLoading {
			g.state = readyState
		}
	case spinner.TickMsg:
		if g.state == loadingState && g.spinner.ID() == msg.ID {
			s, cmd := g.spinner.Update(msg)
			g.spinner = s
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case tickMsg:
		log.Info("Received tickMsg", "msg", msg)
		cmds = append(cmds, g.Tick())
		cmds = append(cmds, g.updateModels(msg))
	default:
		log.Warn("Unhandled message type", "msg", msg)
	}
	active := g.panes[g.activeTab]
	if g.dump != nil {
		log.Debug("Updating active tab", "tab", active.TabName())
	}
	m, cmd := active.Update(msg)
	g.panes[g.activeTab] = m.(common.TabComponent)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Update the status bar on these events
	// Must come after we've updated the active tab
	switch msg.(type) {
	case tabs.ActiveTabMsg, tea.KeyMsg, selector.ActiveMsg, GameMsg, GoBackMsg:
		g.setStatusBarInfo()
	}

	s, cmd := g.statusbar.Update(msg)
	g.statusbar = s.(*statusbar.Model)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return g, tea.Batch(cmds...)
}

func (g Game) View() string {
	wm, hm := g.getMargins()
	hm += g.common.Styles.Tabs.GetHeight() +
		g.common.Styles.Tabs.GetVerticalFrameSize()
	s := g.common.Styles.Repo.Base.
		Width(g.common.Width - wm).
		Height(g.common.Height - hm)
	mainStyle := g.common.Styles.Repo.Body.
		Height(g.common.Height - hm)
	var main string
	var statusbar string
	switch g.state {
	case loadingState:
		main = fmt.Sprintf("%s loading…", g.spinner.View())
		statusbar = g.statusbar.View()
	case readyState:
		main = g.panes[g.activeTab].View()
		statusbar = g.statusbar.View()
	}
	main = g.common.Zone.Mark(
		"game-main",
		mainStyle.Render(main),
	)
	view := lipgloss.JoinVertical(lipgloss.Left,
		g.headerView(),
		g.tabs.View(),
		main,
		statusbar,
	)
	return s.Render(view)
}

func main() {
	// Open or create the log file
	logFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set up the logger to log debug messages to the file
	log.SetOutput(logFile)
	log.SetLevel(log.DebugLevel)
	log.Debug("Starting up")

	// Properly initialize common.Common
	ctx := context.Background()
	renderer := lipgloss.NewRenderer(os.Stdout)
	c := common.NewCommon(ctx, renderer, 0, 0)

	comps := []common.TabComponent{
		NewOverviewModel(c),
		NewBuildingsModel(c),
		NewCapitalModel(c),
		NewWeaponsModel(c),
	}
	g := newGame(c, comps...)
	if _, err := tea.NewProgram(g, tea.WithAltScreen()).Run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func (g *Game) headerView() string {
	log.Debug("Rendering header", "game", g.game)
	truncate := g.common.Renderer.NewStyle().MaxWidth(g.common.Width)
	header := g.game.ProjectName
	if header == "" {
		header = g.game.Name
	}
	header = g.common.Styles.Repo.HeaderName.Render(header)
	desc := strings.TrimSpace(g.game.Description)
	if desc != "" {
		header = lipgloss.JoinVertical(lipgloss.Left,
			header,
			g.common.Styles.Repo.HeaderDesc.Render(desc),
		)
	}
	urlStyle := g.common.Styles.URLStyle.
		Width(g.common.Width - lipgloss.Width(desc) - 1).
		Align(lipgloss.Right)
	var url string
	if cfg := g.common.Config(); cfg != nil {
		url = g.common.CloneCmd(cfg.SSH.PublicURL, g.game.Name)
	}
	url = common.TruncateString(url, g.common.Width-lipgloss.Width(desc)-1)
	url = g.common.Zone.Mark(
		fmt.Sprintf("%s-url", g.game.Name),
		urlStyle.Render(url),
	)

	header = lipgloss.JoinHorizontal(lipgloss.Top, header, url)

	style := g.common.Styles.Repo.Header.Width(g.common.Width)
	return style.Render(
		truncate.Render(header),
	)
}

func goBackCmd() tea.Msg {
	return GoBackMsg{}
}

func (g *Game) debugGameMsg() tea.Msg {
	log.Debug("Debugging GameMsg", "game", g)
	return GameMsg(g)
}

func (g *Game) setStatusBarInfo() {
	active := g.panes[g.activeTab]
	key := g.game.Name
	value := active.StatusBarValue()
	info := active.StatusBarInfo()
	extra := "*"

	g.statusbar.SetStatus(key, value, info, extra)
}

func (g *Game) updateTabComponent(c common.TabComponent, msg tea.Msg) tea.Cmd {
	log.Debug("Updating tab component", "tab", c.TabName())
	cmds := make([]tea.Cmd, 0)
	for i, b := range g.panes {
		if b.TabName() == c.TabName() {
			m, cmd := b.Update(msg)
			g.panes[i] = m.(common.TabComponent)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			break
		}
	}
	return tea.Batch(cmds...)
}

func (g *Game) updateModels(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i, b := range g.panes {
		m, cmd := b.Update(msg)
		g.panes[i] = m.(common.TabComponent)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func switchTabCmd(m common.TabComponent) tea.Cmd {
	return func() tea.Msg {
		return SwitchTabMsg(m)
	}
}

func renderLoading(c common.Common, s spinner.Model) string {
	msg := fmt.Sprintf("%s loading…", s.View())
	return c.Styles.SpinnerContainer.
		Height(c.Height).
		Render(msg)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
