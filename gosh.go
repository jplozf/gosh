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
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"gosh/cmd"
	"gosh/conf"
	"gosh/edit"
	"gosh/fm"
	"gosh/hexedit"
	"gosh/pm"
	"gosh/sq3"
	"gosh/ui"
	"gosh/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	appDir   string
	hostname string
	greeting string
	err      error
)

// ****************************************************************************
// init()
// ****************************************************************************
func init() {
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	cmd.CurrentUser = user.Username
	greeting = cmd.CurrentUser + "@" + hostname + ":"

	cmd.ICmd = 0
	ui.App = tview.NewApplication()
	ui.SetUI(appQuit, greeting)

	userDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	conf.Cwd = userDir
	fm.Hidden = false
	appDir = filepath.Join(userDir, conf.APP_FOLDER)
	if _, err := os.Stat(appDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(appDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	conf.LogFile, err = os.OpenFile(filepath.Join(appDir, conf.LOG_FILE), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	/*
		defer f.Close()

		if _, err = f.WriteString(text); err != nil {
			panic(err)
		}
	*/
	readSettings()
	pm.CurrentView = pm.VIEW_PROCESS
	pm.InitSignals()
	sq3.CurrentDatabaseName = ":memory:"
	sq3.SetSQLMenu()
}

// ****************************************************************************
// main()
// ****************************************************************************
func main() {
	// Main keyboard's events manager
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			cmd.SwitchToHelp()
		case tcell.KeyF2:
			cmd.SwitchToShell()
		case tcell.KeyF3:
			cmd.SwitchToFiles()
		case tcell.KeyF4:
			cmd.SwitchToProcess()
		case tcell.KeyF6:
			cmd.SwitchToEditor()
		case tcell.KeyF9:
			cmd.SwitchToSQLite3()
		case tcell.KeyF12:
			ui.PgsApp.ShowPage("dlgQuit")
		case tcell.KeyCtrlC:
			return nil
		case tcell.KeyCtrlO:
			if ui.CurrentMode == ui.ModeSQLite3 {
				sq3.DoOpenDB(conf.Cwd)
			}
			if ui.CurrentMode == ui.ModeHexEdit {
				hexedit.DoOpen(conf.Cwd)
			}
		case tcell.KeyCtrlS:
			if ui.CurrentMode == ui.ModeSQLite3 {
				sq3.DoCloseDB()
			}
			if ui.CurrentMode == ui.ModeHexEdit {
				hexedit.Close()
			}
		}
		return event
	})

	// Files panel keyboard's events manager
	ui.TblFiles.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			fm.ProceedFileAction()
			return nil
		case tcell.KeyF5:
			fm.RefreshMe()
		case tcell.KeyF8:
			fm.ShowMenu()
			return nil
		case tcell.KeyCtrlS:
			fm.ShowMenuSort()
			return nil
		case tcell.KeyInsert:
			fm.ProceedFileSelect()
			return nil
		case tcell.KeyCtrlA:
			fm.SelectAll()
			return nil
		case tcell.KeyCtrlC:
			fm.DoCopy()
			return nil
		case tcell.KeyCtrlX:
			fm.DoCut()
			return nil
		case tcell.KeyCtrlV:
			fm.DoPaste()
			return nil
		case tcell.KeyDelete:
			fm.DoDelete()
			return nil
		case tcell.KeyTab:
			if ui.TxtPrompt.HasFocus() {
				ui.App.SetFocus(ui.TblFiles)
			} else {
				ui.App.SetFocus(ui.TxtFileInfo)
			}
			return nil
		}
		return event
	})

	// Process panel keyboard's events manager
	ui.TblProcess.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			pm.ProceedProcessAction()
			return nil
		case tcell.KeyF5:
			pm.RefreshMe()
		case tcell.KeyF8:
			pm.ShowMenu()
			return nil
		case tcell.KeyCtrlF:
			pm.DoFindProcess()
			return nil
		case tcell.KeyCtrlS:
			pm.ShowMenuSort()
			return nil
		case tcell.KeyCtrlV:
			pm.SwitchView()
			return nil
		case tcell.KeyTab:
			if ui.TxtPrompt.HasFocus() {
				ui.App.SetFocus(ui.TblProcess)
				return nil
			}
			if ui.TblProcess.HasFocus() {
				ui.App.SetFocus(ui.TblProcUsers)
				return nil
			}
			if ui.TblProcUsers.HasFocus() {
				ui.App.SetFocus(ui.TxtPrompt)
				return nil
			}
		}
		return event
	})

	// TblProcUsers panel keyboard's events manager
	ui.TblProcUsers.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			idx, _ := ui.TblProcUsers.GetSelection()
			pm.ShowProcesses(ui.TblProcUsers.GetCell(idx, 1).Text)
			ui.App.Sync()
			ui.App.SetFocus(ui.TblProcess)
			return nil
		/*
			case tcell.KeyF8:
				fm.ShowMenu()
				return nil
		*/
		case tcell.KeyTab:
			if ui.TblProcUsers.HasFocus() {
				ui.App.SetFocus(ui.TxtProcInfo)
				return nil
			}
		}
		return event
	})

	// ProcInfo keyboard's events manager
	ui.TxtProcInfo.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			ui.App.SetFocus(ui.TxtPrompt)
			return nil
		}
		return event
	})

	// FileInfo keyboard's events manager
	ui.TxtFileInfo.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			ui.App.SetFocus(ui.TxtPrompt)
			return nil
		}
		return event
	})

	// Prompt keyboard's events manager
	ui.TxtPrompt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if ui.CurrentMode == ui.ModeSQLite3 {
				if ui.TxtPrompt.GetText() != "" {
					sq3.Xeq(ui.TxtPrompt.GetText())
				}
			} else {
				if ui.TxtPrompt.GetText() != "" {
					cmd.Xeq(ui.TxtPrompt.GetText())
				}
			}
			return nil
		case tcell.KeyUp:
			if ui.CurrentMode == ui.ModeSQLite3 {
				if len(sq3.ACmd) > 0 {
					if sq3.ICmd < len(sq3.ACmd)-1 {
						sq3.ICmd++
					} else {
						sq3.ICmd = 0
					}
					ui.TxtPrompt.SetText(sq3.ACmd[sq3.ICmd], true)
					ui.TxtPrompt.Select(0, ui.TxtPrompt.GetTextLength())
				}
			} else {
				if len(cmd.ACmd) > 0 {
					if cmd.ICmd < len(cmd.ACmd)-1 {
						cmd.ICmd++
					} else {
						cmd.ICmd = 0
					}
					ui.TxtPrompt.SetText(cmd.ACmd[cmd.ICmd], true)
					ui.TxtPrompt.Select(0, ui.TxtPrompt.GetTextLength())
				}
			}
			return nil
		case tcell.KeyDown:
			if ui.CurrentMode == ui.ModeSQLite3 {
				if len(sq3.ACmd) > 0 {
					if sq3.ICmd > 0 {
						sq3.ICmd--
					} else {
						sq3.ICmd = len(sq3.ACmd) - 1
					}
					ui.TxtPrompt.SetText(sq3.ACmd[sq3.ICmd], true)
					ui.TxtPrompt.Select(0, ui.TxtPrompt.GetTextLength())
				}
			} else {
				if len(cmd.ACmd) > 0 {
					if cmd.ICmd > 0 {
						cmd.ICmd--
					} else {
						cmd.ICmd = len(cmd.ACmd) - 1
					}
					ui.TxtPrompt.SetText(cmd.ACmd[cmd.ICmd], true)
					ui.TxtPrompt.Select(0, ui.TxtPrompt.GetTextLength())
				}
			}
			return nil
		case tcell.KeyTab:
			if ui.CurrentMode == ui.ModeFiles {
				ui.App.SetFocus(ui.TblFiles)
			}
			if ui.CurrentMode == ui.ModeShell {
				ui.App.SetFocus(ui.TxtConsole)
			}
			if ui.CurrentMode == ui.ModeProcess {
				ui.App.SetFocus(ui.TblProcess)
			}
			if ui.CurrentMode == ui.ModeTextEdit {
				ui.App.SetFocus(ui.EdtMain)
			}
			if ui.CurrentMode == ui.ModeSQLite3 {
				ui.App.SetFocus(ui.TblSQLOutput)
			}
			return nil
		}
		return event
	})

	// Console keyboard's events manager
	ui.TxtConsole.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ui.App.SetFocus(ui.TxtPrompt)
			return nil
		}
		return event
	})

	// Editor keyboard's events manager
	ui.EdtMain.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		evkSaveAs := tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModAlt)
		if event.Key() == evkSaveAs.Key() && event.Rune() == evkSaveAs.Rune() && event.Modifiers() == evkSaveAs.Modifiers() {
			edit.SaveFileAs()
			return nil
		}
		switch event.Key() {
		case tcell.KeyCtrlS:
			edit.SaveFile()
			return nil
		case tcell.KeyCtrlN:
			edit.NewFile(conf.Cwd)
			return nil
		case tcell.KeyCtrlT:
			edit.CloseCurrentFile()
			return nil
		case tcell.KeyEsc:
			ui.App.SetFocus(ui.TblOpenFiles)
			return nil
		}
		return event
	})
	ui.TblOpenFiles.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ui.App.SetFocus(ui.TrvExplorer)
			return nil
		case tcell.KeyEnter:
			idx, _ := ui.TblOpenFiles.GetSelection()
			fName := ui.TblOpenFiles.GetCell(idx, 3).Text
			edit.SwitchOpenFile(fName)
			ui.App.SetFocus(ui.EdtMain)
			return nil
		}
		return event
	})
	ui.TrvExplorer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ui.App.SetFocus(ui.TxtPrompt)
			return nil
		}
		return event
	})

	// SQLite3 keyboard's events manager
	ui.TblSQLOutput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF8:
			sq3.ShowMenu()
			return nil
		case tcell.KeyTab:
			ui.App.SetFocus(ui.TblSQLTables)
			return nil
		}
		return event
	})
	ui.TblSQLTables.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ui.App.SetFocus(ui.TrvSQLDatabase)
			return nil
		}
		return event
	})
	ui.TrvSQLDatabase.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ui.App.SetFocus(ui.TxtPrompt)
			return nil
		}
		return event
	})

	// ui.LblKeys.SetText("F1=Help F3=Files F4=Process F12=Exit")
	ui.SetTitle(conf.APP_STRING)
	ui.SetStatus("Welcome.")
	cmd.SwitchToShell()
	welcome()

	go ui.UpdateTime()
	go utils.GetCpuUsage()
	if err := ui.App.SetRoot(ui.PgsApp, true).SetFocus(ui.FlxMain).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// ****************************************************************************
// appQuit()
// appQuit performs some cleanup and saves persistent data before quitting application
// ****************************************************************************
func appQuit() {
	edit.CheckOpenFilesForSaving()
	saveSettings()
	ui.SetStatus("Quitting...")
	ui.App.Stop()
	fmt.Printf("\n👻%s\n\n", conf.APP_STRING)
}

// ****************************************************************************
// readSettings()
// ****************************************************************************
func readSettings() {
	// Read commands history file
	ui.SetStatus("Reading commands history")
	fCmd, err := os.Open(filepath.Join(appDir, conf.HISTORY_CMD_FILE))
	if err != nil {
		return
	}
	defer fCmd.Close()
	sCmd := bufio.NewScanner(fCmd)
	for sCmd.Scan() {
		cmd.ACmd = append(cmd.ACmd, sCmd.Text())
	}
	// Read SQL history file
	ui.SetStatus("Reading SQL history")
	fSQL, err := os.Open(filepath.Join(appDir, conf.HISTORY_SQL_FILE))
	if err != nil {
		return
	}
	defer fSQL.Close()
	sSQL := bufio.NewScanner(fSQL)
	for sSQL.Scan() {
		sq3.ACmd = append(sq3.ACmd, sSQL.Text())
	}
}

// ****************************************************************************
// saveSettings()
// ****************************************************************************
func saveSettings() {
	// Save commands history file
	ui.SetStatus("Saving commands history")
	fCmd, err := os.Create(filepath.Join(appDir, conf.HISTORY_CMD_FILE))
	if err != nil {
		return
	}
	defer fCmd.Close()
	wCmd := bufio.NewWriter(fCmd)
	for _, line := range cmd.ACmd {
		fmt.Fprintln(wCmd, line)
	}
	wCmd.Flush()
	// Save SQL history file
	ui.SetStatus("Saving SQL history")
	fSQL, err := os.Create(filepath.Join(appDir, conf.HISTORY_SQL_FILE))
	if err != nil {
		return
	}
	defer fSQL.Close()
	wSQL := bufio.NewWriter(fSQL)
	for _, line := range sq3.ACmd {
		fmt.Fprintln(wSQL, line)
	}
	wSQL.Flush()
}

// ****************************************************************************
// welcome()
// ****************************************************************************
func welcome() {
	w1 := ":: Welcome to " + conf.APP_STRING + " :"
	w2 := conf.APP_NAME + " version " + conf.APP_VERSION + " - " + conf.APP_URL + "\n"
	os := runtime.GOOS
	if os == "windows" {
		out, err := exec.Command("ver").Output()
		if err == nil {
			w2 = w2 + string(out)
		}

	} else {
		out, err := exec.Command("uname", "-a").Output()
		if err == nil {
			w2 = w2 + string(out)
		}
	}
	ui.LblHostname.SetText("👻" + greeting)
	ui.HeaderConsole(w1)
	ui.OutConsole(w2)
}
