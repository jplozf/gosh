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
	conf.SetLog()
	log.SetOutput(&conf.LogFile)
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
	fm.Cwd = userDir
	fm.Hidden = false
	appDir = filepath.Join(userDir, conf.APP_FOLDER)
	if _, err := os.Stat(appDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(appDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	readSettings()
	pm.CurrentView = pm.VIEW_PROCESS
	pm.InitSignals()
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
			if len(cmd.ACmd) > 0 {
				if cmd.ICmd < len(cmd.ACmd)-1 {
					cmd.ICmd++
				} else {
					cmd.ICmd = 0
				}
				ui.TxtPrompt.SetText(cmd.ACmd[cmd.ICmd], true)
				ui.TxtPrompt.Select(0, ui.TxtPrompt.GetTextLength())
			}
			return nil
		case tcell.KeyDown:
			if len(cmd.ACmd) > 0 {
				if cmd.ICmd > 0 {
					cmd.ICmd--
				} else {
					cmd.ICmd = len(cmd.ACmd) - 1
				}
				ui.TxtPrompt.SetText(cmd.ACmd[cmd.ICmd], true)
				ui.TxtPrompt.Select(0, ui.TxtPrompt.GetTextLength())
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
			edit.NewFile(fm.Cwd)
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
	ui.LblKeys.SetText("F1=Help F3=Files F4=Process F12=Exit")
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
	ui.App.Stop()
	fmt.Printf("\n👻%s\n\n", conf.APP_STRING)
}

// ****************************************************************************
// readSettings()
// ****************************************************************************
func readSettings() {
	// Read history commands file
	file, err := os.Open(filepath.Join(appDir, conf.HISTORY_CMD_FILE))
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd.ACmd = append(cmd.ACmd, scanner.Text())
	}
}

// ****************************************************************************
// saveSettings()
// ****************************************************************************
func saveSettings() {
	// Save history commands file
	file, err := os.Create(filepath.Join(appDir, conf.HISTORY_CMD_FILE))
	if err != nil {
		return
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for _, line := range cmd.ACmd {
		fmt.Fprintln(w, line)
	}
	w.Flush()
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
