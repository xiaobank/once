package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/once/internal/docker"
)

type settingsMenuKeyMap struct {
	Close key.Binding
}

func (k settingsMenuKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Close}
}

func (k settingsMenuKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Close}}
}

var settingsMenuKeys = settingsMenuKeyMap{
	Close: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
}

type settingsMenuItem int

const (
	menuItemApplication settingsMenuItem = iota
	menuItemEmail
	menuItemEnvironment
	menuItemResources
	menuItemCount
)

type SettingsMenuCloseMsg struct{}

type SettingsMenuSelectMsg struct {
	app     *docker.Application
	section SettingsSectionType
}

type SettingsMenu struct {
	app           *docker.Application
	selected      settingsMenuItem
	width, height int
	help          Help
}

func NewSettingsMenu(app *docker.Application) SettingsMenu {
	return SettingsMenu{
		app:      app,
		selected: menuItemApplication,
		help:     NewHelp(),
	}
}

func (m SettingsMenu) Init() tea.Cmd {
	return nil
}

func (m SettingsMenu) Update(msg tea.Msg) (SettingsMenu, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.MouseClickMsg:
		if cmd := m.help.Update(msg, settingsMenuKeys); cmd != nil {
			return m, cmd
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.selected = (m.selected - 1 + menuItemCount) % menuItemCount
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.selected = (m.selected + 1) % menuItemCount
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m, m.selectCurrent()
		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			return m, m.selectSection(SettingsSectionApplication)
		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			return m, m.selectSection(SettingsSectionEmail)
		case key.Matches(msg, key.NewBinding(key.WithKeys("v"))):
			return m, m.selectSection(SettingsSectionEnvironment)
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			return m, m.selectSection(SettingsSectionResources)
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return SettingsMenuCloseMsg{} }
		}
	}

	return m, nil
}

func (m SettingsMenu) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 4)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Primary).
		MarginBottom(1)

	itemStyle := lipgloss.NewStyle()

	selectedStyle := lipgloss.NewStyle().Reverse(true)

	keyStyle := lipgloss.NewStyle().Foreground(Colors.Border)

	title := titleStyle.Render("Settings")

	items := []string{
		m.renderItem("Application", "a", menuItemApplication, itemStyle, selectedStyle, keyStyle),
		m.renderItem("Email", "e", menuItemEmail, itemStyle, selectedStyle, keyStyle),
		m.renderItem("Environment", "v", menuItemEnvironment, itemStyle, selectedStyle, keyStyle),
		m.renderItem("Resources", "r", menuItemResources, itemStyle, selectedStyle, keyStyle),
	}

	helpView := m.help.View(settingsMenuKeys)
	itemWidth := 14 // label padding (13) + shortcut (1)
	helpLine := lipgloss.NewStyle().MarginTop(1).Width(itemWidth).Align(lipgloss.Center).Render(helpView)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		strings.Join(items, "\n"),
		helpLine,
	)

	return boxStyle.Render(content)
}

// Private

func (m SettingsMenu) renderItem(label, shortcut string, item settingsMenuItem, normalStyle, selectedStyle, keyStyle lipgloss.Style) string {
	// Pad to align shortcut keys (longest label is "Environment" at 11 chars)
	padding := strings.Repeat(" ", 13-len(label))
	styledKey := keyStyle.Render(shortcut)

	if m.selected == item {
		return selectedStyle.Render(label) + padding + styledKey
	}
	return normalStyle.Render(label) + padding + styledKey
}

func (m SettingsMenu) selectCurrent() tea.Cmd {
	var section SettingsSectionType
	switch m.selected {
	case menuItemApplication:
		section = SettingsSectionApplication
	case menuItemEmail:
		section = SettingsSectionEmail
	case menuItemEnvironment:
		section = SettingsSectionEnvironment
	case menuItemResources:
		section = SettingsSectionResources
	}
	return m.selectSection(section)
}

func (m SettingsMenu) selectSection(section SettingsSectionType) tea.Cmd {
	return func() tea.Msg {
		return SettingsMenuSelectMsg{app: m.app, section: section}
	}
}
