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
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"gosh/conf"
	"gosh/edit"
	"gosh/fm"
	"gosh/pm"
	"gosh/ui"
	"gosh/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	appDir      string
	aCmd        []string
	iCmd        int
	currentUser string
	hostname    string
	greeting    string
	err         error
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
	currentUser = user.Username
	greeting = currentUser + "@" + hostname + ":"

	iCmd = 0
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
}

// ****************************************************************************
// main()
// ****************************************************************************
func main() {
	// Main keyboard's events manager
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			switchToHelp()
		case tcell.KeyF2:
			switchToShell()
		case tcell.KeyF3:
			switchToFiles()
		case tcell.KeyF4:
			switchToProcess()
		case tcell.KeyF6:
			switchToEditor()
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
		case tcell.KeyCtrlS:
			pm.ShowMenuSort()
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
				ui.App.SetFocus(ui.TxtPrompt)
				return nil
			}
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
			if ui.TxtPrompt.GetText() != "" {
				// TODO : Manage the case we are not in Shell mode
				xeq(ui.TxtPrompt.GetText())
			}
			return nil
		case tcell.KeyUp:
			if len(aCmd) > 0 {
				if iCmd < len(aCmd)-1 {
					iCmd++
				} else {
					iCmd = 0
				}
				ui.TxtPrompt.SetText(aCmd[iCmd], true)
				ui.TxtPrompt.Select(0, ui.TxtPrompt.GetTextLength())
			}
			return nil
		case tcell.KeyDown:
			if len(aCmd) > 0 {
				if iCmd > 0 {
					iCmd--
				} else {
					iCmd = len(aCmd) - 1
				}
				ui.TxtPrompt.SetText(aCmd[iCmd], true)
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

	ui.EdtMain.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			edit.SaveFile()
			return nil
		}
		return event
	})
	ui.LblKeys.SetText("F1=Help F3=Files F4=Process F12=Exit")
	ui.SetTitle(conf.APP_STRING)
	ui.SetStatus("Welcome.")
	switchToShell()
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
	saveSettings()
	ui.App.Stop()
	fmt.Println("👻" + conf.APP_STRING)
}

// ****************************************************************************
// xeq()
// ****************************************************************************
func xeq(c string) {
	// c = " bash -c " + c
	sCmd := strings.Fields(c)
	aCmd = append(aCmd, c)
	iCmd++
	switch sCmd[0] {
	case "cls":
		ui.TxtConsole.SetText("")
	case "quit":
		ui.PgsApp.SwitchToPage("dlgQuit")
	case "shell":
		switchToShell()
	case "files":
		switchToFiles()
	case "process":
		switchToProcess()
	case "editor":
		switchToEditor()
	default:
		switchToShell()
		ui.SetStatus("Running [" + c + "]")
		ui.HeaderConsole(c)

		xCmd := exec.Command(sCmd[0], sCmd[1:]...)
		xCmd.Stdout = io.MultiWriter(&ui.StdoutBuf) // ui.StdoutBuf is displayed through the ui.UpdateTime go routine
		err := xCmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
	}
	ui.TxtPrompt.SetText("", false)
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
		aCmd = append(aCmd, scanner.Text())
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
	for _, line := range aCmd {
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

// ****************************************************************************
// switchToHelp()
// ****************************************************************************
func switchToHelp() {
	ui.SetTitle("Help")
	ui.LblKeys.SetText("F2=Shell F3=Files F4=Process F6=Editor F12=Exit")
	ui.PgsApp.SwitchToPage("help")
}

// ****************************************************************************
// switchToShell()
// ****************************************************************************
func switchToShell() {
	ui.CurrentMode = ui.ModeShell
	ui.SetTitle("Shell")
	ui.LblKeys.SetText("F1=Help F3=Files F4=Process F6=Editor F12=Exit")
	ui.PgsApp.SwitchToPage("main")
}

// ****************************************************************************
// switchToShell()
// ****************************************************************************
func switchToEditor() {
	edit.SwitchToEditor()
}

// ****************************************************************************
// switchToFiles()
// ****************************************************************************
func switchToFiles() {
	ui.CurrentMode = ui.ModeFiles
	fm.SetFilesMenu()
	ui.SetTitle("Files")
	ui.LblKeys.SetText("F1=Help F2=Shell F4=Process F5=Refresh F6=Editor F8=Actions F12=Exit\nIns=Select Ctrl+A=Select/Unselect All Ctrl+C=Copy Ctrl+X=Cut Ctrl+V=Paste Ctrl+S=Sort")
	fm.ShowFiles()
	ui.App.Sync()
	ui.App.SetFocus(ui.TblFiles)
}

// ****************************************************************************
// switchToProcess()
// ****************************************************************************
func switchToProcess() {
	ui.CurrentMode = ui.ModeProcess
	pm.SetProcessMenu()
	ui.SetTitle("Process")
	ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F5=Refresh F6=Editor F8=Actions F12=Exit\nCtrl+F=Find Ctrl+S=Sort")
	pm.ShowProcesses(currentUser)
	ui.App.Sync()
	ui.App.SetFocus(ui.TblProcess)
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
