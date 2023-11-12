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
package ui

import (
	"fmt"
	"gosh/conf"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Fn func()

var (
	lblTime    *tview.TextView
	lblDate    *tview.TextView
	LblKeys    *tview.TextView
	App        *tview.Application
	Flex       *tview.Flex
	TxtPrompt  *tview.TextArea
	TxtConsole *tview.TextView
	lblTitle   *tview.TextView
	lblStatus  *tview.TextView
	lblRC      *tview.TextView
	Pages      *tview.Pages
	modal      *tview.Modal
)

// ****************************************************************************
// setUI()
// setUI defines the user interface's fields
// ****************************************************************************
func SetUI(fQuit Fn) {
	Pages = tview.NewPages()

	lblDate = tview.NewTextView().SetText(currentDateString())
	lblDate.SetBorder(false)

	lblTime = tview.NewTextView().SetText(currentTimeString())
	lblTime.SetBorder(false)

	LblKeys = tview.NewTextView()
	LblKeys.SetBorder(false)
	LblKeys.SetBackgroundColor(tcell.ColorBlack)
	LblKeys.SetTextColor(tcell.ColorLightBlue)

	lblTitle = tview.NewTextView()
	lblTitle.SetBorder(false)
	lblTitle.SetBackgroundColor(tcell.ColorBlack)
	lblTitle.SetTextColor(tcell.ColorGreen)
	lblTitle.SetBorderColor(tcell.ColorDarkGreen)
	lblTitle.SetTextAlign(tview.AlignCenter)

	lblStatus = tview.NewTextView()
	lblStatus.SetBorder(false)
	lblStatus.SetBackgroundColor(tcell.ColorDarkGreen)
	lblStatus.SetTextColor(tcell.ColorWheat)

	lblRC = tview.NewTextView()
	lblRC.SetBorder(false)
	lblRC.SetBackgroundColor(tcell.ColorDarkGreen)
	lblRC.SetTextColor(tcell.ColorWheat)

	TxtPrompt = tview.NewTextArea().SetPlaceholder("Command to run")
	TxtPrompt.SetBorder(false)

	TxtConsole = tview.NewTextView().Clear()
	TxtConsole.SetBorder(true)

	Flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(TxtConsole, 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 3, 1, true).
		AddItem(tview.NewFlex().
			AddItem(lblStatus, 0, 1, false).
			AddItem(lblRC, 5, 0, false), 1, 0, false)

	modal = tview.NewModal().
		SetText("Do you want to quit the application ?").
		AddButtons([]string{"Quit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				fQuit()
			} else {
				Pages.SwitchToPage("main")
			}
		})

	Pages.AddPage("main", Flex, true, true)
	Pages.AddPage("dialog", modal, false, false)
}

// ****************************************************************************
// currentDateString()
// currentDateString returns the current date formatted as a string
// ****************************************************************************
func currentDateString() string {
	d := time.Now()
	return fmt.Sprint(d.Format("02/01/2006"))
}

// ****************************************************************************
// currentTimeString()
// currentTimeString returns the current time formatted as a string
// ****************************************************************************
func currentTimeString() string {
	t := time.Now()
	return fmt.Sprint(t.Format("15:04:05"))
}

// ****************************************************************************
// updateTime()
// updateTime is the go routine which refresh the time and date
// ****************************************************************************
func UpdateTime() {
	for {
		time.Sleep(500 * time.Millisecond)
		App.QueueUpdateDraw(func() {
			lblDate.SetText(currentDateString())
			lblTime.SetText(currentTimeString())
		})
	}
}

// ****************************************************************************
// setTitle()
// setTitle displays the title centered
// ****************************************************************************
func SetTitle(t string) {
	lblTitle.SetText(t)
}

// ****************************************************************************
// setStatus()
// setStatus displays the status message during a specific time
// ****************************************************************************
func SetStatus(t string) {
	lblStatus.SetText(t)
	DurationOfTime := time.Duration(conf.STATUS_MESSAGE_DURATION) * time.Second
	f := func() {
		lblStatus.SetText("")
	}
	time.AfterFunc(DurationOfTime, f)
}

// ****************************************************************************
// outConsole()
// ****************************************************************************
func OutConsole(cmd string, out string) {
	TxtConsole.SetText(TxtConsole.GetText(true) + "> " + cmd + ":\n" + out + "\n")
	TxtConsole.ScrollToEnd()
}
