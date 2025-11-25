package model

import (
	"fmt"
	"slices"

	"github.com/AdityaKK0407/sentryvault/internal/cipher"
	"github.com/AdityaKK0407/sentryvault/internal/database"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	bolt "go.etcd.io/bbolt"
)

type state uint8

const (
	tableDetails state = iota
	addDetails
	updateDetails
	removeDetails
)

type DetailsModel struct {
	tableView   table.Model
	keyInput    textinput.Model
	valueInput  textinput.Model
	help        help.Model
	state       state
	Entry       string
	db          *bolt.DB
	cipherKey32 []byte
	cipherKey64 []byte
}

func (m DetailsModel) setTableRows() (DetailsModel, error) {
	pairs, err := database.RetrieveAll(m.db, []byte(m.Entry))
	if err != nil {
		return m, err
	}

	var rows []table.Row
	for _, pair := range pairs {
		val, err := cipher.DecryptAESGCM(m.cipherKey32, pair[1])
		if err != nil {
			return m, err
		}
		rows = append(rows, table.Row{
			string(pair[0]),
			string(val),
		})
	}
	m.tableView.SetRows(rows)
	return m, nil
}

func (m DetailsModel) selectBoundsCheck() bool {
	if m.tableView.Cursor() >= 0 && m.tableView.Cursor() < len(m.tableView.Rows()) {
		return true
	}
	return false
}

func initialEntryDetailsModel(db *bolt.DB, cipherKey32, cipherKey64 []byte) DetailsModel {
	cols := []table.Column{
		{Title: "Key", Width: 35},
		{Title: "Value", Width: 35},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
		table.WithHeight(10),
		table.WithFocused(true),
	)
	t.Focus()

	keyInput := textinput.New()
	keyInput.Prompt = "Enter Key: "
	keyInput.Width = 20

	valueInput := textinput.New()
	valueInput.Prompt = "Enter your value: "
	valueInput.Width = 20

	return DetailsModel{
		tableView:   t,
		keyInput:    keyInput,
		valueInput:  valueInput,
		help:        help.New(),
		Entry:       "",
		db:          db,
		cipherKey32: cipherKey32,
		cipherKey64: cipherKey64,
	}
}

func (m DetailsModel) Init() tea.Cmd {
	return nil
}

func (m DetailsModel) Update(msg tea.Msg) (DetailsModel, tea.Cmd) {
	var cmd tea.Cmd
	var commands []tea.Cmd
	kb := keybindings()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, kb.Quit):
			return m, tea.Quit
		case key.Matches(msg, kb.Escape):
			if m.state == tableDetails {
				return m, func() tea.Msg {
					return returnEntryMsg{}
				}
			} else {
				m.keyInput.Reset()
				m.valueInput.Reset()
				m.keyInput.Blur()
				m.valueInput.Blur()
				m.tableView.Focus()
				m.state = tableDetails
			}
		case key.Matches(msg, kb.Enter):
			if m.state == addDetails {
				if m.keyInput.Focused() {
					m.keyInput.Blur()
					m.valueInput.Focus()
				} else if m.valueInput.Focused() {
					keyEntry := m.keyInput.Value()
					value := m.valueInput.Value()
					cipherValue, err := cipher.EncryptAESGCM(m.cipherKey32, []byte(value))
					if err != nil {
						return m, func() tea.Msg {
							return errMsg{Err: err}
						}
					}
					if err = database.Insert(m.db, []byte(m.Entry), []byte(keyEntry), cipherValue); err != nil {
						return m, func() tea.Msg {
							return errMsg{Err: err}
						}
					}
					rows := m.tableView.Rows()
					rows = append(rows, table.Row{
						keyEntry,
						value,
					})
					m.tableView.SetRows(rows)
					m.keyInput.Reset()
					m.keyInput.Blur()
					m.valueInput.Reset()
					m.valueInput.Blur()
					m.tableView.Focus()
					m.state = tableDetails
				}
			} else if m.state == updateDetails {
				keyEntry := m.tableView.SelectedRow()[0]
				value := m.valueInput.Value()
				cipherValue, err := cipher.EncryptAESGCM(m.cipherKey32, []byte(value))
				if err != nil {
					return m, func() tea.Msg {
						return errMsg{Err: err}
					}
				}
				if err = database.Insert(m.db, []byte(m.Entry), []byte(keyEntry), cipherValue); err != nil {
					return m, func() tea.Msg {
						return errMsg{Err: err}
					}
				}
				rows := m.tableView.Rows()
				index := m.tableView.Cursor()
				rows[index][1] = value
				m.tableView.SetRows(rows)
				m.keyInput.Reset()
				m.keyInput.Blur()
				m.valueInput.Reset()
				m.valueInput.Blur()
				m.tableView.Focus()
				m.state = tableDetails
			}
		case key.Matches(msg, kb.Add):
			if m.state == tableDetails {
				m.tableView.Blur()
				m.keyInput.Focus()
				m.state = addDetails
				return m, nil
			}
		case key.Matches(msg, kb.Update):
			if m.state == tableDetails && m.selectBoundsCheck() {
				m.tableView.Blur()
				m.valueInput.Focus()
				m.state = updateDetails
				return m, nil
			}
		case key.Matches(msg, kb.Remove):
			if m.state == tableDetails && m.selectBoundsCheck() {
				m.tableView.Blur()
				m.state = removeDetails
			}
		case key.Matches(msg, kb.Confirm):
			if m.state == removeDetails {
				entry := m.tableView.SelectedRow()[0]
				if err := database.Remove(m.db, []byte(m.Entry), []byte(entry)); err != nil {
					return m, func() tea.Msg {
						return errMsg{Err: err}
					}
				}
				index := m.tableView.Cursor()
				m.tableView.SetRows(slices.Delete(m.tableView.Rows(), index, index+1))
			}
			fallthrough
		case key.Matches(msg, kb.Cancel):
			if m.state == removeDetails {
				m.tableView.Focus()
				m.state = tableDetails
			}
		}
	}
	if m.state == tableDetails {
		m.tableView, cmd = m.tableView.Update(msg)
		commands = append(commands, cmd)
	} else if m.state == addDetails {
		m.keyInput, cmd = m.keyInput.Update(msg)
		commands = append(commands, cmd)
		m.valueInput, cmd = m.valueInput.Update(msg)
		commands = append(commands, cmd)
	} else if m.state == updateDetails {
		m.valueInput, cmd = m.valueInput.Update(msg)
		commands = append(commands, cmd)
	}

	return m, tea.Batch(commands...)
}

func (m DetailsModel) View() string {
	s := fmt.Sprintf("Entry: %s\n\n", m.Entry)
	s += m.tableView.View()

	switch m.state {
	case addDetails:
		s += fmt.Sprintf("\n\n%s", m.keyInput.View())
		s += fmt.Sprintf("\n%s", m.valueInput.View())
	case updateDetails:
		s += fmt.Sprintf("\n\n%s", m.valueInput.View())
	case removeDetails:
		row := m.tableView.SelectedRow()
		s += fmt.Sprintf("\n\nDelete Key: %s, Value: %s?\n", row[0], row[1])
		s += fmt.Sprintf("[y] Yes  [n] No\n\n")
	case tableDetails:
	default:
	}

	s += fmt.Sprintf("\n\n%s\n", m.help.View(keybindings()))

	return s
}
