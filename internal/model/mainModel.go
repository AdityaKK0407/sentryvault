package model

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	bolt "go.etcd.io/bbolt"
)

type modelState uint8

const (
	EntryList modelState = iota
	EntryDetails
)

type MainModel struct {
	state            modelState
	entryListState   EntryModel
	entryDetailState DetailsModel
	db               *bolt.DB
	cipherKey        []byte
	Err              error
}

type selectEntryMsg struct {
	Entry string
}

type returnEntryMsg struct{}

type errMsg struct {
	Err error
}

func InitialMainModel(db *bolt.DB, cipherKey32, cipherKey64 []byte) *MainModel {
	return &MainModel{
		state:            EntryList,
		entryListState:   initialEntryListModel(db, cipherKey64),
		entryDetailState: initialEntryDetailsModel(db, cipherKey32, cipherKey64),
		db:               db,
		cipherKey:        cipherKey32,
		Err:              nil,
	}
}

type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Tab     key.Binding
	Add     key.Binding
	Update  key.Binding
	Remove  key.Binding
	Escape  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
	Quit    key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Enter,
		k.Tab,
		k.Add,
		k.Remove,
		k.Escape,
		k.Quit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Add, k.Remove, k.Escape, k.Quit},
	}
}

func keybindings() KeyMap {
	return KeyMap{
		Up:      key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:    key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Tab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "move back")),
		Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Update:  key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "update")),
		Remove:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "remove")),
		Escape:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "escape add/update/remove model")),
		Confirm: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
		Cancel:  key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "cancel")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.entryListState.Init(),
		m.entryDetailState.Init(),
	)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case selectEntryMsg:
		m.state = EntryDetails
		m.entryDetailState.Entry = msg.Entry
		var err error
		m.entryDetailState, err = m.entryDetailState.setTableRows()
		if err != nil {
			m.Err = err
			return m, tea.Quit
		}
	case returnEntryMsg:
		m.state = EntryList
	case errMsg:
		m.Err = msg.Err
		return m, tea.Quit
	case tea.KeyMsg:
		switch m.state {
		case EntryList:
			m.entryListState, cmd = m.entryListState.Update(msg)
		case EntryDetails:
			m.entryDetailState, cmd = m.entryDetailState.Update(msg)
		}
	}
	return m, cmd
}

func (m MainModel) View() string {
	switch m.state {
	case EntryList:
		return m.entryListState.View()
	case EntryDetails:
		fallthrough
	default:
		return m.entryDetailState.View()
	}
}
