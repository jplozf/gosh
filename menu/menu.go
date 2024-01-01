// ****************************************************************************
//
//	 _____ _____ _____ _____
//	|   __|     |   __|  |  |
//	|  |  |  |  |__   |     |
//	|_____|_____|_____|__|__|
//
// ****************************************************************************
// G O S H   -   Copyright © JPL 2023
// ****************************************************************************
package menu

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"gosh/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type MenuItem struct {
	Name    string
	Label   string
	Done    func()
	Enabled bool
}

type Menu struct {
	*tview.Table
	title  string
	items  []MenuItem
	parent string
	focus  tview.Primitive
	width  int
	height int
}

// ****************************************************************************
// New() MenuItem
// ****************************************************************************
func (mi *MenuItem) New(name string, label string, done func(), enabled bool) *MenuItem {
	mi = &MenuItem{
		Name:    name,
		Label:   label,
		Done:    done,
		Enabled: enabled,
	}
	return mi
}

// ****************************************************************************
// New() Menu
// ****************************************************************************
func (m *Menu) New(title string, parent string, focus tview.Primitive) *Menu {
	m = &Menu{
		Table:  tview.NewTable(),
		title:  title,
		parent: parent,
		focus:  focus,
	}
	return m
}

// ****************************************************************************
// AddItem() Menu
// ****************************************************************************
func (m *Menu) AddItem(name string, label string, event func(), enabled bool) {
	var item *MenuItem
	item = item.New(name, label, event, enabled)
	m.items = append(m.items, *item)
}

// ****************************************************************************
// SetEnabled() Menu
// ****************************************************************************
func (m *Menu) SetEnabled(miName string, e bool) {
	for index, item := range m.items {
		if item.Name == miName {
			m.items[index].Enabled = e
		}
	}
	m.refresh()
}

// ****************************************************************************
// SetLabel() Menu
// ****************************************************************************
func (m *Menu) SetLabel(miName string, label string) {
	for index, item := range m.items {
		if item.Name == miName {
			m.items[index].Label = label
		}
	}
	m.refresh()
}

// ****************************************************************************
// refresh() Menu
// ****************************************************************************
func (m *Menu) refresh() {
	m.width = 0
	m.height = len(m.items)
	m.Table.SetBorder(true)
	m.Table.SetTitle(m.title)
	m.Table.SetSelectable(true, false)
	m.Table.SetBackgroundColor(tcell.ColorBlue)
	for i, item := range m.items {
		item.Label = "  " + item.Label + "  "
		if item.Enabled {
			m.Table.SetCell(i, 0, tview.NewTableCell(item.Label).SetTextColor(tcell.ColorYellow))
		} else {
			m.Table.SetCell(i, 0, tview.NewTableCell(item.Label).SetTextColor(tcell.ColorGray))
		}
		if len(item.Label) > m.width {
			m.width = len(item.Label)
		}
	}
	m.width = m.width + 2
	m.height = m.height + 2
}

// ****************************************************************************
// Popup()
// ****************************************************************************
func (m *Menu) Popup() tview.Primitive {
	m.refresh()
	m.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			idx, _ := m.Table.GetSelection()
			if m.items[idx].Enabled {
				ui.PgsApp.SwitchToPage(m.parent)
				ui.App.SetFocus(m.focus)
				m.items[idx].Done()
			}
			return nil
		case tcell.KeyEsc:
			ui.PgsApp.SwitchToPage(m.parent)
			ui.App.SetFocus(m.focus)
			return nil
		}
		return event
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(m, m.height, 1, true).
			AddItem(nil, 0, 1, false), m.width, 1, true).
		AddItem(nil, 0, 1, false)
}
