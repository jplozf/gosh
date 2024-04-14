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
	"bufio"
	"bytes"
	"fmt"
	"gosh/conf"
	"gosh/edit"
	"gosh/fm"
	"gosh/help"
	"gosh/hexedit"
	"gosh/pm"
	"gosh/sq3"
	"gosh/ui"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-cmd/cmd"
)

var (
	ACmd        []string
	ICmd        int
	CurrentUser string
)

type output struct {
	buf   *bytes.Buffer
	lines []string
	*sync.Mutex
}

func newOutput() *output {
	return &output{
		buf:   &bytes.Buffer{},
		lines: []string{},
		Mutex: &sync.Mutex{},
	}
}

// io.Writer interface is only this method
func (rw *output) Write(p []byte) (int, error) {
	rw.Lock()
	defer rw.Unlock()
	return rw.buf.Write(p) // and bytes.Buffer implements it, too
}

func (rw *output) Lines() []string {
	rw.Lock()
	defer rw.Unlock()
	// Scanners are io.Readers which effectively destroy the buffer by reading
	// to EOF. So once we scan the buf to lines, the buf is empty again.
	s := bufio.NewScanner(rw.buf)
	for s.Scan() {
		rw.lines = append(rw.lines, s.Text())
	}
	return rw.lines
}

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
			// SwitchToShell()
			ui.AddNewScreen(ui.ModeShell, nil, nil)
		case "!file":
			// SwitchToFiles()
			ui.AddNewScreen(ui.ModeFiles, fm.SelfInit, nil)
		case "!proc":
			// SwitchToProcess()
			ui.AddNewScreen(ui.ModeProcess, pm.SelfInit, CurrentUser)
		case "!edit":
			// SwitchToEditor()
			ui.AddNewScreen(ui.ModeTextEdit, edit.SelfInit, nil)
		case "!sql":
			// SwitchToSQLite3()
			ui.AddNewScreen(ui.ModeSQLite3, sq3.SelfInit, nil)
		case "!help":
			// SwitchToHelp()
			ui.AddNewScreen(ui.ModeHelp, help.SelfInit, nil)
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
			ui.SetStatus(fmt.Sprintf("Running [%s]", c))
			ui.HeaderConsole(c)

			cmdOptions := cmd.Options{
				Buffered:  false,
				Streaming: true,
			}

			xCmd := cmd.NewCmdOptions(cmdOptions, sCmd[0], sCmd[1:]...)
			doneChan := make(chan struct{})
			go func() {
				defer close(doneChan)
				// Done when both channels have been closed
				// https://dave.cheney.net/2013/04/30/curious-channels
				for xCmd.Stdout != nil || xCmd.Stderr != nil {
					select {
					case line, open := <-xCmd.Stdout:
						if !open {
							xCmd.Stdout = nil
							continue
						}
						ui.TxtConsole.SetText(ui.TxtConsole.GetText(false) + line + "\n")
					case line, open := <-xCmd.Stderr:
						if !open {
							xCmd.Stderr = nil
							continue
						}
						ui.TxtConsole.SetText(ui.TxtConsole.GetText(false) + "[yellow]" + line + "[white]\n")
					}
				}
			}()

			// Run and wait for Cmd to return
			status := <-xCmd.Start()
			ui.LblPID.SetText(fmt.Sprintf("PID=%d", status.PID))

			// Wait for goroutine to print everything
			<-doneChan
			ui.LblRC.SetText(fmt.Sprintf("RC=%d", status.Exit))

		}
	}
	ui.TxtPrompt.SetText("", false)
}

/*
// ****************************************************************************
// SwitchToHelp()
// ****************************************************************************
func SwitchToHelp() {
	ui.SetTitle("Help")
	help.SetHelp()
	ui.LblKeys.SetText(conf.FKEY_LABELS)
	ui.PgsApp.SwitchToPage("help")
	ui.App.SetFocus(ui.TxtHelp)
}
*/

/*
// ****************************************************************************
// SwitchToShell()
// ****************************************************************************
func SwitchToShell() {
	ui.CurrentMode = ui.ModeShell
	ui.SetTitle("Shell")
	ui.LblKeys.SetText(conf.FKEY_LABELS)
	ui.PgsApp.SwitchToPage("shell")
}
*/

/*
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
*/

/*
// ****************************************************************************
// SwitchToFiles()
// ****************************************************************************
func SwitchToFiles() {
	ui.CurrentMode = ui.ModeFiles
	fm.SetFilesMenu()
	ui.SetTitle("Files")
	ui.LblKeys.SetText(conf.FKEY_LABELS + "\nDel=Delete Ins=Select Ctrl+A=Select/Unselect All Ctrl+C=Copy Ctrl+X=Cut Ctrl+V=Paste Ctrl+S=Sort")
	fm.ShowFiles()
	ui.App.Sync()
	ui.App.SetFocus(ui.TblFiles)
}
*/

/*
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
					ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
					ui.PgsApp.SwitchToPage("sqlite3")
					ui.App.SetFocus(ui.TxtPrompt)
				} else {
					ui.CurrentMode = ui.ModeSQLite3
					ui.SetTitle("SQLite3")
					ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
					ui.PgsApp.SwitchToPage("sqlite3")
					ui.App.SetFocus(ui.TxtPrompt)
					ui.SetStatus(err.Error())
				}
			} else {
				// attach the targeted database to the current database
				sq3.DoExec(fmt.Sprintf("attach database '%s' as %s", fName, utils.FilenameWithoutExtension(filepath.Base(fName))))
				ui.CurrentMode = ui.ModeSQLite3
				ui.SetTitle("SQLite3")
				ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
				ui.PgsApp.SwitchToPage("sqlite3")
				ui.App.SetFocus(ui.TxtPrompt)
			}
		} else {
			ui.CurrentMode = ui.ModeSQLite3
			ui.SetTitle("SQLite3")
			ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
			ui.PgsApp.SwitchToPage("sqlite3")
			ui.App.SetFocus(ui.TxtPrompt)
		}
	} else {
		ui.CurrentMode = ui.ModeSQLite3
		ui.SetTitle("SQLite3")
		ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
		ui.PgsApp.SwitchToPage("sqlite3")
		ui.App.SetFocus(ui.TxtPrompt)
	}
}
*/

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
			ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
			ui.PgsApp.SwitchToPage("hexedit")
			ui.App.SetFocus(ui.TxtPrompt)
		} else {
			ui.CurrentMode = ui.ModeHexEdit
			ui.SetTitle("HexEdit")
			ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
			ui.PgsApp.SwitchToPage("hexedit")
			ui.App.SetFocus(ui.TxtPrompt)
		}
	} else {
		ui.CurrentMode = ui.ModeHexEdit
		ui.SetTitle("HexEdit")
		ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
		ui.PgsApp.SwitchToPage("hexedit")
		ui.App.SetFocus(ui.TxtPrompt)
	}
}

/*
// ****************************************************************************
// SwitchToProcess()
// ****************************************************************************
func SwitchToProcess() {
	ui.CurrentMode = ui.ModeProcess
	pm.SetProcessMenu()
	ui.SetTitle("Process")
	ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+F=Find Ctrl+S=Sort Ctrl+V=Switch View")
	pm.ShowProcesses(CurrentUser)
	ui.App.Sync()
	ui.App.SetFocus(ui.TblProcess)
}
*/
