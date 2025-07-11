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
	"strings"
	"sync"
	syscall "syscall"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/rivo/tview"
)

var (
	ACmd        []string
	ICmd        int
	CurrentUser string
	currentCmd  *cmd.Cmd
	currentCmdPGID int
)

// StopCurrentCommand stops the currently running command.
func StopCurrentCommand() {
	if currentCmdPGID != 0 {
		ui.SetStatus("Attempting to interrupt command...")
		ui.SetStatus(fmt.Sprintf("Sending SIGINT to PGID %d", currentCmdPGID))
        err := syscall.Kill(-currentCmdPGID, syscall.SIGINT) // Send to process group
        if err != nil {
            conf.LogFile.WriteString(fmt.Sprintf("cmd.go: Error sending SIGINT to PGID %d: %v\n", currentCmdPGID, err))
        }
        time.Sleep(100 * time.Millisecond) // Give process time to react
        currentCmdPGID = 0
		ui.SetStatus("Command interrupted.")
	} else {
		ui.SetStatus("No active command to interrupt.")
	}
}

type output struct {
	buf   *bytes.Buffer
	lines []string
	Mutex *sync.Mutex
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
	rw.Mutex.Lock()
	defer rw.Mutex.Unlock()
	return rw.buf.Write(p) // and bytes.Buffer implements it, too
}

func (rw *output) Lines() []string {
	rw.Mutex.Lock()
	defer rw.Mutex.Unlock()
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

			// Run command in a goroutine to prevent blocking the UI
            go func() {
                conf.LogFile.WriteString(fmt.Sprintf("cmd.go: Starting command goroutine for: %s\n", c))
                currentCmd = xCmd // Store the command for interruption

                statusChan := xCmd.Start() // Start the command

                time.Sleep(50 * time.Millisecond) // Give process time to start
                status := xCmd.Status()
                currentCmdPGID = status.PID // Use PID as PGID for now
                conf.LogFile.WriteString(fmt.Sprintf("cmd.go: Command started, PID: %d, PGID: %d (initial from status)\n", status.PID, currentCmdPGID))

                // Goroutine to handle output streaming
                doneChan := make(chan struct{})
                go func() {
                    defer close(doneChan)

                    for {
                        select {
                        case line, open := <-xCmd.Stdout:
                            if !open {
                                xCmd.Stdout = nil
                                conf.LogFile.WriteString("cmd.go: Stdout channel closed.\n")
                                continue
                            }
                            conf.LogFile.WriteString(fmt.Sprintf("cmd.go: Stdout line: %s\n", line))
                            ui.App.QueueUpdateDraw(func() { // Update UI on main thread
                                ui.LblPID.SetText(fmt.Sprintf("PID=%d", currentCmdPGID)) // Use currentCmdPGID for display
                                ui.TxtConsole.Write([]byte(tview.TranslateANSI(line) + "\n"))
                                ui.TxtConsole.ScrollToEnd()
                            })

                        case line, open := <-xCmd.Stderr:
                            if !open {
                                xCmd.Stderr = nil
                                conf.LogFile.WriteString("cmd.go: Stderr channel closed.\n")
                                continue
                            }
                            conf.LogFile.WriteString(fmt.Sprintf("cmd.go: Stderr line: %s\n", line))
                            ui.App.QueueUpdateDraw(func() { // Update UI on main thread
                                ui.LblPID.SetText(fmt.Sprintf("PID=%d", currentCmdPGID)) // Use currentCmdPGID for display
                                ui.TxtConsole.Write([]byte("[yellow]" + tview.TranslateANSI(line) + "[white]\n"))
                                ui.TxtConsole.ScrollToEnd()
                            })
                        case <-time.After(100 * time.Millisecond): // Periodically check if statusChan is closed
                            if statusChan == nil {
                                break // Exit loop if statusChan is closed
                            }
                        }
                        if xCmd.Stdout == nil && xCmd.Stderr == nil && statusChan == nil {
                            break // All channels closed, exit loop
                        }
                    }
                    conf.LogFile.WriteString("cmd.go: Output streaming goroutine finished.\n")
                }()

                // Wait for output streaming to finish
                <-doneChan

                // Wait for the command to finish and get its final status
                finalStatus := <-statusChan // This will get the final status
                conf.LogFile.WriteString("cmd.go: doneChan closed, processing final status.\n")

                // Job's done !
                conf.LogFile.WriteString("cmd.go: Executing UI update for summary.\n")
                ui.TxtPath.SetText(conf.Cwd)
                rc := finalStatus.Exit
                summary := fmt.Sprintf("\n[yellow]Runtime for PID %d is %f seconds. Return Code: %d[white]\n", finalStatus.PID, finalStatus.Runtime, rc)
                conf.LogFile.WriteString(fmt.Sprintf("cmd.go: Writing summary to console: %s\n", summary))
                ui.TxtConsole.Write([]byte(summary))
                if rc != 0 {
                    ui.LblRC.SetText(fmt.Sprintf("[#FF0000]RC=%d", rc))
                } else {
                    ui.LblRC.SetText(fmt.Sprintf("[#F5DEB3]RC=%d", rc))
                }
                ui.LblPID.SetText(fmt.Sprintf("%.2fs", finalStatus.Runtime))
                ui.JobsDone()
                ui.App.Draw() // Force redraw
                time.Sleep(50 * time.Millisecond) // Give UI time to update
                conf.LogFile.WriteString(fmt.Sprintf("cmd.go: Command goroutine finished for: %s\n", c))
            }()
		}
	}
	ui.TxtPrompt.SetText("", false)
}
