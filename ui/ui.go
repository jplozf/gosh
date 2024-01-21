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
	"bytes"
	"fmt"
	"gosh/conf"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
)

type Fn func()

type Mode int64

const (
	ModeShell Mode = iota
	ModeFiles
	ModeTextEdit
	ModeHexEdit
	ModeProcess
	ModeNetwork
	ModeSQLite3
)

var (
	CurrentMode    Mode
	lblTime        *tview.TextView
	lblDate        *tview.TextView
	LblKeys        *tview.TextView
	App            *tview.Application
	FlxMain        *tview.Flex
	FlxFiles       *tview.Flex
	FlxProcess     *tview.Flex
	FlxHelp        *tview.Flex
	FlxEditor      *tview.Flex
	FlxSQL         *tview.Flex
	FlxHexEdit     *tview.Flex
	TxtPrompt      *tview.TextArea
	TxtConsole     *tview.TextView
	TxtFileInfo    *tview.TextView
	TxtProcInfo    *tview.TextView
	TxtHelp        *tview.TextView
	lblTitle       *tview.TextView
	lblStatus      *tview.TextView
	LblHostname    *tview.TextView
	LblRC          *tview.TextView
	PgsApp         *tview.Pages
	dlgQuit        *tview.Modal
	TblFiles       *tview.Table
	TblProcess     *tview.Table
	TxtPath        *tview.TextView
	TxtProcess     *tview.TextView
	FrmFileInfo    *tview.TextView
	TblProcUsers   *tview.Table
	TxtSelection   *tview.TextView
	StdoutBuf      bytes.Buffer
	EdtMain        *femto.View
	TxtEditName    *tview.TextView
	TblOpenFiles   *tview.Table
	TrvExplorer    *tview.TreeView
	TxtSQLName     *tview.TextView
	TblSQLOutput   *tview.Table
	TblSQLTables   *tview.Table
	TrvSQLDatabase *tview.TreeView
	TxtHexName     *tview.TextView
	TblHexEdit     *tview.Table
)

// ****************************************************************************
// init()
// ****************************************************************************
func init() {
	log.SetOutput(&conf.LogFile)
}

// ****************************************************************************
// setUI()
// setUI defines the user interface's fields
// ****************************************************************************
func SetUI(fQuit Fn, hostname string) {
	PgsApp = tview.NewPages()

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

	LblRC = tview.NewTextView()
	LblRC.SetBorder(false)
	LblRC.SetBackgroundColor(tcell.ColorDarkGreen)
	LblRC.SetTextColor(tcell.ColorWheat)

	LblHostname = tview.NewTextView()
	LblHostname.SetBorder(false)
	LblHostname.SetBackgroundColor(tcell.ColorDarkGreen)
	LblHostname.SetTextColor(tcell.ColorBlack)

	TxtPrompt = tview.NewTextArea().SetPlaceholder("Command to run")
	TxtPrompt.SetBorder(false)

	TxtHelp = tview.NewTextView().Clear()
	TxtHelp.SetBorder(true)
	TxtHelp.SetDynamicColors(true)

	TxtConsole = tview.NewTextView().Clear()
	TxtConsole.SetBorder(true)
	TxtConsole.SetDynamicColors(true)

	FrmFileInfo = tview.NewTextView()
	FrmFileInfo.SetBorder(true)
	FrmFileInfo.SetDynamicColors(true)
	FrmFileInfo.SetTitle("Infos")

	TxtFileInfo = tview.NewTextView().Clear()
	TxtFileInfo.SetBorder(true)
	TxtFileInfo.SetDynamicColors(true)
	TxtFileInfo.SetTitle("Preview")
	TxtFileInfo.SetWrap(false)
	TxtFileInfo.SetScrollable(true)

	TxtSelection = tview.NewTextView()
	TxtSelection.SetBorder(true)
	TxtSelection.SetDynamicColors(true)
	TxtSelection.SetTitle("Selection")

	TblFiles = tview.NewTable()
	TblFiles.SetBorder(true)
	TblFiles.SetSelectable(true, false)

	TxtPath = tview.NewTextView()
	TxtPath.Clear()
	TxtPath.SetBorder(true)

	TblProcUsers = tview.NewTable()
	TblProcUsers.SetBorder(true)
	TblProcUsers.SetTitle("Users")
	TblProcUsers.SetSelectable(true, false)

	TxtProcInfo = tview.NewTextView().Clear()
	TxtProcInfo.SetBorder(true)
	TxtProcInfo.SetDynamicColors(true)
	TxtProcInfo.SetTitle("Details")
	TxtProcInfo.SetWrap(false)
	TxtProcInfo.SetScrollable(true)

	TblProcess = tview.NewTable()
	TblProcess.SetBorder(true)
	TblProcess.SetSelectable(true, false)

	TxtProcess = tview.NewTextView()
	TxtProcess.Clear()
	TxtProcess.SetBorder(true)
	TxtProcess.SetDynamicColors(true)

	buffer := femto.NewBufferFromString(string("content"), "./dummy")
	EdtMain = femto.NewView(buffer)
	EdtMain.SetBorder(true)
	TxtEditName = tview.NewTextView()
	TxtEditName.Clear()
	TxtEditName.SetBorder(true)
	TblOpenFiles = tview.NewTable()
	TblOpenFiles.SetBorder(true)
	TblOpenFiles.SetSelectable(true, false)
	TblOpenFiles.SetTitle("Open Files")
	TrvExplorer = tview.NewTreeView()
	TrvExplorer.SetBorder(true)
	TrvExplorer.SetTitle("Explorer")

	TxtSQLName = tview.NewTextView()
	TxtSQLName.Clear()
	TxtSQLName.SetBorder(true)
	TxtSQLName.SetDynamicColors(true)
	TblSQLOutput = tview.NewTable()
	TblSQLOutput.SetBorder(true)
	TblSQLOutput.SetSelectable(true, true)
	TblSQLOutput.SetTitle("Output")
	TblSQLTables = tview.NewTable()
	TblSQLTables.SetBorder(true)
	TblSQLTables.SetSelectable(true, false)
	TblSQLTables.SetTitle("Tables")
	TrvSQLDatabase = tview.NewTreeView()
	TrvSQLDatabase.SetBorder(true)
	TrvSQLDatabase.SetTitle("Database")

	TxtHexName = tview.NewTextView()
	TxtHexName.Clear()
	TxtHexName.SetBorder(true)
	TxtHexName.SetDynamicColors(true)
	TblHexEdit = tview.NewTable()
	TblHexEdit.SetBorder(true)
	TblHexEdit.SetSelectable(true, true)
	TblHexEdit.SetTitle("Hexa View")

	//*************************************************************************
	// Main Layout (Shell)
	//*************************************************************************
	FlxMain = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(TxtConsole, 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 2, 1, true).
		AddItem(tview.NewFlex().
			AddItem(LblHostname, len(hostname)+3, 0, false).
			AddItem(lblStatus, 0, 1, false).
			AddItem(LblRC, 5, 0, false), 1, 0, false)

	//*************************************************************************
	// Help Layout
	//*************************************************************************
	FlxHelp = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(TxtHelp, 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 2, 1, true).
		AddItem(tview.NewFlex().
			AddItem(LblHostname, len(hostname)+3, 0, false).
			AddItem(lblStatus, 0, 1, false).
			AddItem(LblRC, 5, 0, false), 1, 0, false)

	//*************************************************************************
	// Files Layout
	//*************************************************************************
	FlxFiles = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TxtPath, 3, 0, false).
				AddItem(TblFiles, 0, 1, true), 0, 2, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(FrmFileInfo, 9, 0, false).
				AddItem(TxtFileInfo, 0, 1, false).
				AddItem(TxtSelection, 5, 0, false), 0, 1, false), 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 2, 1, true).
		AddItem(tview.NewFlex().
			AddItem(LblHostname, len(hostname)+3, 0, false).
			AddItem(lblStatus, 0, 1, false).
			AddItem(LblRC, 5, 0, false), 1, 0, false)

	TblFiles.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			TblFiles.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		TblFiles.GetCell(row, column).SetTextColor(tcell.ColorRed)
		TblFiles.SetSelectable(false, false)
	})

	//*************************************************************************
	// Process Layout
	//*************************************************************************
	FlxProcess = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TxtProcess, 3, 0, false).
				AddItem(TblProcess, 0, 1, true), 0, 2, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TblProcUsers, 12, 0, false).
				AddItem(TxtProcInfo, 0, 1, false), 0, 1, false), 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 2, 1, true).
		AddItem(tview.NewFlex().
			AddItem(LblHostname, len(hostname)+3, 0, false).
			AddItem(lblStatus, 0, 1, false).
			AddItem(LblRC, 5, 0, false), 1, 0, false)

	//*************************************************************************
	// Editor Layout
	//*************************************************************************
	FlxEditor = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TxtEditName, 3, 0, false).
				AddItem(EdtMain, 0, 1, true), 0, 2, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TblOpenFiles, 12, 0, false).
				AddItem(TrvExplorer, 0, 1, false), 0, 1, false), 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 2, 1, true).
		AddItem(tview.NewFlex().
			AddItem(LblHostname, len(hostname)+3, 0, false).
			AddItem(lblStatus, 0, 1, false).
			AddItem(LblRC, 5, 0, false), 1, 0, false)

	//*************************************************************************
	// SQLite3 Layout
	//*************************************************************************
	FlxSQL = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TxtSQLName, 3, 0, false).
				AddItem(TblSQLOutput, 0, 1, true), 0, 2, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TblSQLTables, 12, 0, false).
				AddItem(TrvSQLDatabase, 0, 1, false), 0, 1, false), 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 2, 1, true).
		AddItem(tview.NewFlex().
			AddItem(LblHostname, len(hostname)+3, 0, false).
			AddItem(lblStatus, 0, 1, false).
			AddItem(LblRC, 5, 0, false), 1, 0, false)

	//*************************************************************************
	// HexaEditor Layout
	//*************************************************************************
	FlxHexEdit = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(lblDate, 10, 0, false).
			AddItem(lblTitle, 0, 1, false).
			AddItem(lblTime, 8, 0, false), 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TxtHexName, 3, 0, false).
				AddItem(TblHexEdit, 0, 1, true), 0, 2, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(TblSQLTables, 12, 0, false).
				AddItem(TrvSQLDatabase, 0, 1, false), 0, 1, false), 0, 1, false).
		AddItem(LblKeys, 2, 1, false).
		AddItem(TxtPrompt, 2, 1, true).
		AddItem(tview.NewFlex().
			AddItem(LblHostname, len(hostname)+3, 0, false).
			AddItem(lblStatus, 0, 1, false).
			AddItem(LblRC, 5, 0, false), 1, 0, false)

	//*************************************************************************
	// Misc
	//*************************************************************************
	dlgQuit = tview.NewModal().
		SetText("Do you want to quit the application ?").
		AddButtons([]string{"Quit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				fQuit()
			} else {
				// TODO : Get the real previous page and go back to it
				SetTitle("Shell")
				LblKeys.SetText("F1=Help F3=Files F12=Exit")
				PgsApp.SwitchToPage("main")
			}
		})

	PgsApp.AddPage("main", FlxMain, true, true)
	PgsApp.AddPage("help", FlxHelp, true, false)
	PgsApp.AddPage("files", FlxFiles, true, false)
	PgsApp.AddPage("process", FlxProcess, true, false)
	PgsApp.AddPage("editor", FlxEditor, true, false)
	PgsApp.AddPage("sqlite3", FlxSQL, true, false)
	PgsApp.AddPage("hexedit", FlxHexEdit, true, false)
	PgsApp.AddPage("dlgQuit", dlgQuit, false, false)
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
			TxtConsole.SetText(TxtConsole.GetText(false) + string(StdoutBuf.Bytes()))
			StdoutBuf.Reset()
			// TxtConsole.SetText(string(StderrBuf.Bytes()))
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
// GetTitle()
// setTitle displays the title centered
// ****************************************************************************
func GetTitle() string {
	return lblTitle.GetText(true)
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
	// TODO : Send status string to log file
	// log.Fatal(t)
}

// ****************************************************************************
// HeaderConsole()
// ****************************************************************************
func HeaderConsole(cmd string) {
	TxtConsole.SetText(TxtConsole.GetText(false) + "\n[red]⯈ " + cmd + ":\n[white]")
	TxtConsole.ScrollToEnd()
}

// ****************************************************************************
// outConsole()
// ****************************************************************************
func OutConsole(out string) {
	TxtConsole.SetText(TxtConsole.GetText(false) + "[white]" + out + "\n")
	TxtConsole.ScrollToEnd()
	App.Sync()
}

// ****************************************************************************
// DisplayMap()
// ****************************************************************************
func DisplayMap(tv *tview.TextView, m map[string]string) {
	// out := tv.GetText(true)
	out := ""
	maxi := 0
	for key := range m {
		if len(key) > maxi {
			maxi = len(key)
		}
	}
	// create slice and store keys
	fields := make([]string, 0, len(m))
	for k := range m {
		fields = append(fields, k)
	}

	// sort the slice by keys
	sort.Strings(fields)

	// iterate by sorted keys
	for _, field := range fields {
		out = out + "[red]" + field[2:] + strings.Repeat(" ", maxi-len(field)) + "[white]  " + m[field] + "\n"
	}
	tv.SetText(out)
}

// ****************************************************************************
// PromptInput()
// ****************************************************************************
func PromptInput(msg string, choice string) {
	TxtPrompt.SetText(msg, true)
}
