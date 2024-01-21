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
package cmd

import (
	"fmt"
	"gosh/conf"
	"gosh/edit"
	"gosh/fm"
	"gosh/help"
	"gosh/hexedit"
	"gosh/pm"
	"gosh/sq3"
	"gosh/ui"
	"gosh/utils"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

var (
	ACmd        []string
	ICmd        int
	CurrentUser string
)

// ****************************************************************************
// xeq()
// ****************************************************************************
func Xeq(c string) {
	sCmd := strings.Fields(c)
	ACmd = append(ACmd, c)
	ICmd++
	if sCmd[0][0] == '!' {
		xCmd := sCmd[0] + "     "
		xCmd = xCmd[:5]
		xCmd = strings.TrimSpace(xCmd)
		switch xCmd {
		case "!quit", "!exit":
			ui.PgsApp.SwitchToPage("dlgQuit")
		case "!shel":
			SwitchToShell()
		case "!file":
			SwitchToFiles()
		case "!proc":
			SwitchToProcess()
		case "!edit":
			SwitchToEditor()
		case "!sql":
			SwitchToSQLite3()
		case "!help":
			SwitchToHelp()
		case "!hex":
			SwitchToHexEdit()
		default:
			ui.SetStatus(fmt.Sprintf("Invalid command %s", sCmd[0]))
		}
	} else {
		switch sCmd[0] {
		case "cls":
			ui.TxtConsole.SetText("")
		default:
			SwitchToShell()
			ui.SetStatus(fmt.Sprintf("Running [%s]", c))
			ui.HeaderConsole(c)

			xCmd := exec.Command(sCmd[0], sCmd[1:]...)
			xCmd.Stdout = io.MultiWriter(&ui.StdoutBuf) // ui.StdoutBuf is displayed through the ui.UpdateTime go routine
			err := xCmd.Run()
			if err != nil {
				ui.SetStatus(err.Error())
			}
		}
	}
	ui.TxtPrompt.SetText("", false)
}

// ****************************************************************************
// SwitchToHelp()
// ****************************************************************************
func SwitchToHelp() {
	ui.SetTitle("Help")
	help.SetHelp()
	ui.LblKeys.SetText("F2=Shell F3=Files F4=Process F6=Editor F9=SQLite3 F12=Exit")
	ui.PgsApp.SwitchToPage("help")
	ui.App.SetFocus(ui.TxtHelp)
}

// ****************************************************************************
// SwitchToShell()
// ****************************************************************************
func SwitchToShell() {
	ui.CurrentMode = ui.ModeShell
	ui.SetTitle("Shell")
	ui.LblKeys.SetText("F1=Help F3=Files F4=Process F6=Editor F9=SQLite3 F12=Exit")
	ui.PgsApp.SwitchToPage("main")
}

// ****************************************************************************
// SwitchToEditor()
// ****************************************************************************
func SwitchToEditor() {
	if ui.CurrentMode == ui.ModeFiles {
		idx, _ := ui.TblFiles.GetSelection()
		fName := filepath.Join(conf.Cwd, strings.TrimSpace(ui.TblFiles.GetCell(idx, 2).Text))
		mtype := utils.GetMimeType(fName)
		if len(mtype) > 3 {
			if mtype[:4] == "text" {
				edit.SwitchToEditor(fName)
			} else {
				edit.NewFileOrLastFile(conf.Cwd)
			}
		} else {
			edit.NewFileOrLastFile(conf.Cwd)
		}
	} else {
		edit.NewFileOrLastFile(conf.Cwd)
	}
}

// ****************************************************************************
// SwitchToFiles()
// ****************************************************************************
func SwitchToFiles() {
	ui.CurrentMode = ui.ModeFiles
	fm.SetFilesMenu()
	ui.SetTitle("Files")
	ui.LblKeys.SetText("F1=Help F2=Shell F4=Process F5=Refresh F6=Editor F8=Actions F9=SQLite3 F12=Exit\nDel=Delete Ins=Select Ctrl+A=Select/Unselect All Ctrl+C=Copy Ctrl+X=Cut Ctrl+V=Paste Ctrl+S=Sort")
	fm.ShowFiles()
	ui.App.Sync()
	ui.App.SetFocus(ui.TblFiles)
}

// ****************************************************************************
// SwitchToSQLite3()
// ****************************************************************************
func SwitchToSQLite3() {
	if ui.CurrentMode == ui.ModeFiles {
		idx, _ := ui.TblFiles.GetSelection()
		fName := filepath.Join(conf.Cwd, strings.TrimSpace(ui.TblFiles.GetCell(idx, 2).Text))
		xtype, _ := mimetype.DetectFile(fName)
		if strings.HasSuffix(xtype.String(), "sqlite3") {
			// Is there an open database ?
			if sq3.CurrentDB == nil {
				// no, then open the targeted database
				err := sq3.OpenDB(fName)
				if err == nil {
					ui.CurrentMode = ui.ModeSQLite3
					ui.SetTitle("SQLite3")
					ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F12=Exit\nCtrl+O=Open Ctrl+S=Save")
					ui.PgsApp.SwitchToPage("sqlite3")
					ui.App.SetFocus(ui.TxtPrompt)
				} else {
					ui.CurrentMode = ui.ModeSQLite3
					ui.SetTitle("SQLite3")
					ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F12=Exit\nCtrl+O=Open Ctrl+S=Save")
					ui.PgsApp.SwitchToPage("sqlite3")
					ui.App.SetFocus(ui.TxtPrompt)
					ui.SetStatus(err.Error())
				}
			} else {
				// attach the targeted database to the current database
				sq3.DoExec(fmt.Sprintf("attach database '%s' as %s", fName, utils.FilenameWithoutExtension(filepath.Base(fName))))
				ui.CurrentMode = ui.ModeSQLite3
				ui.SetTitle("SQLite3")
				ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F12=Exit\nCtrl+O=Open Ctrl+S=Save")
				ui.PgsApp.SwitchToPage("sqlite3")
				ui.App.SetFocus(ui.TxtPrompt)
			}
		} else {
			ui.CurrentMode = ui.ModeSQLite3
			ui.SetTitle("SQLite3")
			ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F12=Exit\nCtrl+O=Open Ctrl+S=Save")
			ui.PgsApp.SwitchToPage("sqlite3")
			ui.App.SetFocus(ui.TxtPrompt)
		}
	} else {
		ui.CurrentMode = ui.ModeSQLite3
		ui.SetTitle("SQLite3")
		ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F12=Exit\nCtrl+O=Open Ctrl+S=Save")
		ui.PgsApp.SwitchToPage("sqlite3")
		ui.App.SetFocus(ui.TxtPrompt)
	}
}

// ****************************************************************************
// SwitchToHexEdit()
// ****************************************************************************
func SwitchToHexEdit() {
	if ui.CurrentMode == ui.ModeFiles {
		if hexedit.CurrentHexFile == "" {
			idx, _ := ui.TblFiles.GetSelection()
			fName := filepath.Join(conf.Cwd, strings.TrimSpace(ui.TblFiles.GetCell(idx, 2).Text))
			hexedit.Open(fName)
			ui.CurrentMode = ui.ModeHexEdit
			ui.SetTitle("HexEdit")
			ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F9=SQLite3 F12=Exit\nCtrl+O=Open Ctrl+S=Save")
			ui.PgsApp.SwitchToPage("hexedit")
			ui.App.SetFocus(ui.TxtPrompt)
		} else {
			ui.CurrentMode = ui.ModeHexEdit
			ui.SetTitle("HexEdit")
			ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F9=SQLite3 F12=Exit\nCtrl+O=Open Ctrl+S=Save")
			ui.PgsApp.SwitchToPage("hexedit")
			ui.App.SetFocus(ui.TxtPrompt)
		}
	} else {
		ui.CurrentMode = ui.ModeHexEdit
		ui.SetTitle("HexEdit")
		ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F5=Refresh F6=Editor F8=Actions F9=SQLite3 F12=Exit\nCtrl+O=Open Ctrl+S=Save")
		ui.PgsApp.SwitchToPage("hexedit")
		ui.App.SetFocus(ui.TxtPrompt)
	}
}

// ****************************************************************************
// SwitchToProcess()
// ****************************************************************************
func SwitchToProcess() {
	ui.CurrentMode = ui.ModeProcess
	pm.SetProcessMenu()
	ui.SetTitle("Process")
	ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F5=Refresh F6=Editor F8=Actions F9=SQLite3 F12=Exit\nCtrl+F=Find Ctrl+S=Sort Ctrl+V=Switch View")
	pm.ShowProcesses(CurrentUser)
	ui.App.Sync()
	ui.App.SetFocus(ui.TblProcess)
}
