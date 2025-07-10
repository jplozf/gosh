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
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/go-cmd/cmd"
	"github.com/rivo/tview"
)

var (
	ACmd        []string
	ICmd        int
	CurrentUser string
	currentCmd  *cmd.Cmd
)

// StopCurrentCommand stops the currently running command.
func StopCurrentCommand() {
	if currentCmd != nil {
		currentCmd.Stop()
		currentCmd = nil
		ui.SetStatus("Command interrupted.")
	}
}

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
		case "!quit", "!exit", "!bye":
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
			// SwitchToHexEdit()
			ui.AddNewScreen(ui.ModeHexEdit, hexedit.SelfInit, nil)
		default:
			ui.SetStatus(fmt.Sprintf("Invalid command %s", sCmd[0]))
		}
	} else {
		switch sCmd[0] {
		case "cls":
			ui.TxtConsole.SetText("")
		default:
			ui.PleaseWait()
			ui.HeaderConsole(c)

			cmdOptions := cmd.Options{
				Buffered:  false,
				Streaming: true,
			}

			xCmd := cmd.NewCmdOptions(cmdOptions, sCmd[0], sCmd[1:]...)
			xCmd.Dir = conf.Cwd
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
						ui.LblPID.SetText(fmt.Sprintf("PID=%d", xCmd.Status().PID))
						ui.TxtConsole.SetText(ui.TxtConsole.GetText(false) + tview.TranslateANSI(line) + "\n")
						ui.App.ForceDraw()
					case line, open := <-xCmd.Stderr:
						if !open {
							xCmd.Stderr = nil
							continue
						}
						ui.LblPID.SetText(fmt.Sprintf("PID=%d", xCmd.Status().PID))
						ui.TxtConsole.SetText(ui.TxtConsole.GetText(false) + "[yellow]" + tview.TranslateANSI(line) + "[white]\n")
						ui.App.ForceDraw()
					}
				}
				conf.Cwd = getWorkingDirectoryOfPID(xCmd.Status().PID)
			}()

			// Run and wait for Cmd to return
			currentCmd = xCmd
			status := <-xCmd.Start()

			// Wait for goroutine to print everything
			<-doneChan

			// Job's done !
			ui.TxtPath.SetText(conf.Cwd)
			ui.TxtConsole.SetText(ui.TxtConsole.GetText(false) + "\n[yellow]" + fmt.Sprintf("Runtime for PID %d is %f seconds.", xCmd.Status().PID, xCmd.Status().Runtime) + "[white]\n")
			rc := status.Exit
			if rc != 0 {
				ui.LblRC.SetText(fmt.Sprintf("[#FF0000]RC=%d", rc))
			} else {
				ui.LblRC.SetText(fmt.Sprintf("[#F5DEB3]RC=%d", rc))
			}
			ui.LblPID.SetText(fmt.Sprintf("%.2fs", xCmd.Status().Runtime))
			ui.JobsDone()
		}
	}
	ui.TxtPrompt.SetText("", false)
}

// ****************************************************************************
// getWorkingDirectoryOfPID()
// ****************************************************************************
func getWorkingDirectoryOfPID(pid int) string {
	cmd := exec.Command("lsof", "-a", "-d", "cwd", "-p", strconv.Itoa(pid))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		ui.SetStatus(err.Error())
	} else {
		ui.SetStatus(outb.String())
	}
	return outb.String()
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

/*
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
*/

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
