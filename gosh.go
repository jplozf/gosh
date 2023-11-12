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
package main

import (
	"os/exec"

	"gosh/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const APP_STRING = "Gosh © JPL 2023"

// ****************************************************************************
// main()
// ****************************************************************************
func main() {
	ui.App = tview.NewApplication()
	ui.SetUI(appQuit)

	// Main keyboard's events manager
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF4:
			ui.Pages.SwitchToPage("dialog")
		case tcell.KeyCtrlC:
			return nil
		}
		return event
	})

	// Prompt keyboard's events manager
	ui.TxtPrompt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			xeq()
			return nil
		}
		return event
	})

	ui.LblKeys.SetText("F4=Exit")
	ui.SetTitle(APP_STRING)
	ui.SetStatus("Welcome.")

	go ui.UpdateTime()
	if err := ui.App.SetRoot(ui.Pages, true).SetFocus(ui.Flex).Run(); err != nil {
		panic(err)
	}
}

// ****************************************************************************
// appQuit()
// appQuit performs some cleanup and saves persistent data before quitting application
// ****************************************************************************
func appQuit() {
	ui.App.Stop()
}

// ****************************************************************************
// xeq()
// ****************************************************************************
func xeq() {
	ui.SetStatus("Running [" + ui.TxtPrompt.GetText() + "]")
	switch cmd := ui.TxtPrompt.GetText(); cmd {
	case "cls":
		ui.TxtConsole.SetText("")
	case "quit":
		ui.Pages.SwitchToPage("dialog")
	default:
		out, err := exec.Command(cmd).Output()
		ui.OutConsole(cmd, string(out))
		if err != nil {
			ui.OutConsole(cmd, string(err.Error()))
			// lblRC.SetText(
		}
	}
	ui.TxtPrompt.SetText("", false)
}
