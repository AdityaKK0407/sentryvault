package model

import (
	"fmt"
	"slices"

	"github.com/AdityaKK0407/sentryvault/internal/database"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	bolt "go.etcd.io/bbolt"
)

type entryListState uint8

const (
	tableEntry entryListState = iota
	addEntry
	removeEntry
)

type EntryModel struct {
	tableView   table.Model
	inputField  textinput.Model
	help        help.Model
	state       entryListState
	db          *bolt.DB
	cipherKey64 []byte
	message     string
	messageErr  bool
}

func (m EntryModel) selectBoundsCheck() bool {
	if m.tableView.Cursor() >= 0 && m.tableView.Cursor() < len(m.tableView.Rows()) {
		return true
	}
	return false
}

func initialEntryListModel(db *bolt.DB, cipherKey64 []byte) EntryModel {
	cols := []table.Column{
		{Title: "Entries", Width: 50},
	}

	entries, err := database.GetEntries(db)
	if err != nil {
		return EntryModel{}
	}

	var rows []table.Row
	for _, entry := range entries {
		rows = append(rows, table.Row{string(entry)})
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(10),
		table.WithFocused(true),
	)
	t.Focus()

	input := textinput.New()
	input.Width = 50
	input.Prompt = "Entry Name: "

	return EntryModel{
		tableView:   t,
		inputField:  input,
		help:        help.New(),
		state:       tableEntry,
		db:          db,
		cipherKey64: cipherKey64,
	}
}

func (m EntryModel) Init() tea.Cmd {
	return nil
}

func (m EntryModel) Update(msg tea.Msg) (EntryModel, tea.Cmd) {
	var cmd tea.Cmd
	var commands []tea.Cmd
	kb := keybindings()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, kb.Quit):
			return m, tea.Quit
		case key.Matches(msg, kb.Escape):
			m.inputField.Reset()
			m.inputField.Blur()
			m.tableView.Focus()
			m.state = tableEntry
		case key.Matches(msg, kb.Enter):
			switch {
			case m.tableView.Focused():
				if m.selectBoundsCheck() {
					row := m.tableView.SelectedRow()
					return m, func() tea.Msg {
						return selectEntryMsg{Entry: row[0]}
					}
				} else {
					m.message = "Invalid Row selected"
					m.messageErr = true
				}
			case m.inputField.Focused():
				entry := m.inputField.Value()
				if entry != "" {
					if err := database.CreateEntry(m.db, []byte(entry)); err != nil {
						return m, func() tea.Msg {
							return errMsg{Err: err}
						}
					}
					rows := m.tableView.Rows()
					rows = append(rows, table.Row{entry})
					m.tableView.SetRows(rows)
					m.tableView.Focus()
					m.inputField.Blur()
					m.inputField.Reset()
					m.state = tableEntry
					m.message = fmt.Sprintf("New Entry \"%s\" added", entry)
					m.messageErr = false
				}
			}
		case key.Matches(msg, kb.Add):
			if m.state == tableEntry {
				m.tableView.Blur()
				m.inputField.Focus()
				m.state = addEntry
				return m, nil
			}
		case key.Matches(msg, kb.Remove):
			if m.state == tableEntry && m.selectBoundsCheck() {
				m.tableView.Blur()
				m.state = removeEntry
			}
		case key.Matches(msg, kb.Confirm):
			if m.state == removeEntry {
				entry := m.tableView.SelectedRow()[0]
				if err := database.RemoveEntry(m.db, []byte(entry)); err != nil {
					return m, func() tea.Msg {
						return errMsg{Err: err}
					}
				}
				index := m.tableView.Cursor()
				m.tableView.SetRows(slices.Delete(m.tableView.Rows(), index, index+1))
				m.message = fmt.Sprintf("Removed entry \"%s\"", entry)
				m.messageErr = false
			}
			fallthrough
		case key.Matches(msg, kb.Cancel):
			if m.state == removeEntry {
				m.tableView.Focus()
				m.state = tableEntry
			}
		}
	}

	if m.state == tableEntry {
		m.tableView, cmd = m.tableView.Update(msg)
		commands = append(commands, cmd)
	} else if m.state == addEntry {
		m.inputField, cmd = m.inputField.Update(msg)
		commands = append(commands, cmd)
	}
	return m, tea.Batch(commands...)
}

func (m EntryModel) View() string {
	s := tableStyle.Render(m.tableView.View())

	switch m.state {
	case addEntry:
		s += fmt.Sprintf("\n\n%s", m.inputField.View())
	case removeEntry:
		s += fmt.Sprintf("\n\nDelete Entry: %s?\n", m.tableView.SelectedRow()[0])
		s += fmt.Sprintf("[y] Yes  [n] No\n\n")
	case tableEntry:
	default:
	}

	//if m.messageErr {
	//	s += errMessageStyle.Render(m.message)
	//} else {
	//	s += successMessageStyle.Render(m.message)
	//}

	s += fmt.Sprintf("\n\n%s\n", m.help.View(keybindings()))

	return s
}
