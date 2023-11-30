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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"gosh/conf"
	"gosh/fm"
	"gosh/ui"
	"gosh/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type selecao struct {
	fName string
	fSize int64
	fType string
}

var (
	appDir string
	aCmd   []string
	iCmd   int
	sel    []selecao
)

// ****************************************************************************
// init()
// ****************************************************************************
func init() {
	file, err := os.OpenFile("gosh.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)

	iCmd = 0
	ui.App = tview.NewApplication()
	ui.SetUI(appQuit)

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
			proceedFileAction()
			return nil
		case tcell.KeyF8:
			fm.ShowMenu()
			return nil
		case tcell.KeyInsert:
			proceedFileSelect()
			return nil
		case tcell.KeyTab:
			if ui.TxtPrompt.HasFocus() {
				ui.App.SetFocus(ui.TblFiles)
			} else {
				ui.App.SetFocus(ui.TxtPrompt)
			}
			return nil
		}
		return event
	})

	// Prompt keyboard's events manager
	ui.TxtPrompt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if ui.TxtPrompt.GetText() != "" {
				// TODO : Manage the case we are in Files mode
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

	ui.LblKeys.SetText("F1=Help F3=Files F4=Process F12=Exit")
	ui.SetTitle(conf.APP_STRING)
	ui.SetStatus("Welcome.")
	switchToShell()
	ui.OutConsole(welcome())

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
	fmt.Println(conf.APP_STRING)
}

// ****************************************************************************
// xeq()
// ****************************************************************************
func xeq(c string) {
	cmd := strings.Fields(c)
	aCmd = append(aCmd, c)
	iCmd++
	switch cmd[0] {
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
	default:
		switchToShell()
		ui.SetStatus("Running [" + c + "]")
		out, err := exec.Command(cmd[0], cmd[1:]...).Output()
		if err != nil {
			ui.OutConsole(c, string(err.Error())+"\n")
		} else {
			ui.OutConsole(c, string(out))
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
	ui.LblKeys.SetText("F2=Shell F3=Files F4=Process F12=Exit")
	ui.PgsApp.SwitchToPage("help")
}

// ****************************************************************************
// switchToShell()
// ****************************************************************************
func switchToShell() {
	ui.CurrentMode = ui.ModeShell
	ui.SetTitle("Shell")
	ui.LblKeys.SetText("F1=Help F3=Files F4=Process F12=Exit")
	ui.PgsApp.SwitchToPage("main")
}

// ****************************************************************************
// switchToFiles()
// ****************************************************************************
func switchToFiles() {
	ui.CurrentMode = ui.ModeFiles
	fm.SetFilesMenu()
	ui.SetTitle("Files")
	ui.LblKeys.SetText("F1=Help F2=Shell F4=Process F8=Actions F12=Exit\nIns=Select Ctrl+C=Copy Ctrl+V=Paste")
	sel = nil
	ui.TxtSelection.Clear()
	ui.PgsApp.SwitchToPage("files")
	files, err := os.ReadDir(fm.Cwd)
	if err != nil {
		ui.SetStatus(err.Error())
	}

	ui.TblFiles.Clear()
	ui.TxtFileInfo.Clear()
	iStart := 0
	if fm.Cwd != "/" {
		ui.TblFiles.SetCell(0, 0, tview.NewTableCell("   "))
		ui.TblFiles.SetCell(0, 1, tview.NewTableCell("..").SetTextColor(tcell.ColorYellow))
		ui.TblFiles.SetCell(0, 2, tview.NewTableCell("<UP>"))
		ui.TblFiles.SetCell(0, 3, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 4, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 5, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 6, tview.NewTableCell(" "))
		iStart = 1
	}
	ui.TxtPath.SetText(fm.Cwd)
	for i, file := range files {
		if !fm.Hidden && file.Name()[0] == '.' {
			continue
		}
		ui.TblFiles.SetCell(i+iStart, 0, tview.NewTableCell("   "))
		ui.TblFiles.SetCell(i+iStart, 1, tview.NewTableCell(file.Name()).SetTextColor(tcell.ColorYellow))
		fi, err := file.Info()
		if err == nil {
			ui.TblFiles.SetCell(i+iStart, 2, tview.NewTableCell(fi.ModTime().String()[0:19]))
			if fi.IsDir() {
				ui.TblFiles.SetCell(i+iStart, 3, tview.NewTableCell("  FOLDER"))
				ui.TblFiles.SetCell(0, 6, tview.NewTableCell(" "))
			} else {
				if fi.Mode().String()[0] == 'L' {
					ui.TblFiles.SetCell(i+iStart, 3, tview.NewTableCell("  LINK"))
					lnk, err := os.Readlink(filepath.Join(fm.Cwd, ui.TblFiles.GetCell(i+iStart, 1).Text))
					if err == nil {
						ui.TblFiles.SetCell(i+iStart, 6, tview.NewTableCell(lnk))
					} else {
						ui.TblFiles.SetCell(i+iStart, 6, tview.NewTableCell(err.Error()))
					}
				} else {
					ui.TblFiles.SetCell(i+iStart, 3, tview.NewTableCell("  FILE"))
					ui.TblFiles.SetCell(0, 6, tview.NewTableCell(" "))
				}
			}
			ui.TblFiles.SetCell(i+iStart, 4, tview.NewTableCell(fi.Mode().String()))
			ui.TblFiles.SetCell(i+iStart, 5, tview.NewTableCell(strconv.FormatInt(fi.Size(), 10)).SetAlign(tview.AlignRight))
		}
	}
	ui.TblFiles.Select(0, 0)
	ui.App.Sync()
	ui.App.SetFocus(ui.TblFiles)
}

// ****************************************************************************
// switchToProcess()
// ****************************************************************************
func switchToProcess() {
	ui.CurrentMode = ui.ModeProcess
	ui.SetTitle("Process")
	ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F12=Exit")
	sel = nil
	ui.TxtSelection.Clear()
	ui.PgsApp.SwitchToPage("process")
	ui.TxtProcess.SetText(fmt.Sprintf("CPU usage is %.2f%%\n", utils.CpuUsage))
}

// ****************************************************************************
// proceedFileAction()
// ****************************************************************************
func proceedFileAction() {
	idx, _ := ui.TblFiles.GetSelection()
	// TODO : manage LINK
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	if targetType == "LINK" {
		targetType = "FILE"
		fName := filepath.Join(fm.Cwd, ui.TblFiles.GetCell(idx, 1).Text)
		fFile, err := os.Readlink(fName)
		if err == nil {
			info, err := os.Stat(fFile)
			if err == nil {
				if info.IsDir() {
					targetType = "FOLDER"
				}
			} else {
				ui.SetStatus(err.Error())
			}
		} else {
			ui.SetStatus(err.Error())
		}
	}
	if targetType == "FILE" { // or type(readlink)==file
		fName := filepath.Join(fm.Cwd, ui.TblFiles.GetCell(idx, 1).Text)
		ui.FrmFileInfo.Clear()

		size, _ := strconv.ParseFloat(ui.TblFiles.GetCell(idx, 5).Text, 64)
		if size <= conf.HASH_THRESHOLD_SIZE {
			infos := map[string]string{
				"0Name":        ui.TblFiles.GetCell(idx, 1).Text,
				"1Change Date": ui.TblFiles.GetCell(idx, 2).Text,
				"2Access":      ui.TblFiles.GetCell(idx, 4).Text,
				"3Size":        ui.TblFiles.GetCell(idx, 5).Text + " Bytes (" + utils.HumanFileSize(size) + ")",
				"4Mime Type":   utils.GetMimeType(fName),
			}
			ui.DisplayMap(ui.FrmFileInfo, infos)

			f, err := os.OpenFile(fName, os.O_RDONLY, os.ModePerm)
			if err != nil {
				ui.SetStatus(err.Error())
			}
			defer f.Close()

			ui.TxtFileInfo.Clear()
			if utils.IsTextFile(fName) {
				reader := bufio.NewReader(f)
				characters := make([]byte, conf.FILE_MAX_PREVIEW)
				_, err := reader.Read(characters)
				if err != nil {
					ui.SetStatus(err.Error())
				} else {
					ui.TxtFileInfo.SetText(string(characters) + "\n")
				}
			} else {
				sha, err := utils.GetSha256(fName)
				if err != nil {
					ui.SetStatus(err.Error())
				} else {
					ui.TxtFileInfo.SetText("SHA256 : \n" + sha[0:31] + "\n" + sha[32:63])
				}
			}
		} else {
			infos := map[string]string{
				"0Name":        ui.TblFiles.GetCell(idx, 1).Text,
				"1Change Date": ui.TblFiles.GetCell(idx, 2).Text,
				"2Access":      ui.TblFiles.GetCell(idx, 4).Text,
				"3Size":        ui.TblFiles.GetCell(idx, 5).Text + " Bytes (" + utils.HumanFileSize(size) + ")",
			}
			ui.DisplayMap(ui.FrmFileInfo, infos)
			ui.TxtFileInfo.SetText("VERY BIG FILE, can't display a preview.")
		}
	} else { //  or type(readlink)==folder
		fm.Cwd = filepath.Join(fm.Cwd, ui.TblFiles.GetCell(idx, 1).Text)
		ui.FrmFileInfo.Clear()
		nFiles, nFolders, err := utils.NumberOfFilesAndFolders(fm.Cwd)
		if err != nil {
			ui.SetStatus(err.Error())
		}
		infos := map[string]string{
			"0Name":    fm.Cwd,
			"1Files":   strconv.Itoa(nFiles),
			"2Folders": strconv.Itoa(nFolders),
		}
		ui.DisplayMap(ui.FrmFileInfo, infos)
		switchToFiles()
	}
	ui.App.Sync()
}

// ****************************************************************************
// proceedFileSelect()
// ****************************************************************************
func proceedFileSelect() {
	idx, _ := ui.TblFiles.GetSelection()
	if strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text) == "FILE" {
		if ui.TblFiles.GetCell(idx, 0).Text == "   " {
			// SELECT FILE
			fName := filepath.Join(fm.Cwd, ui.TblFiles.GetCell(idx, 1).Text)
			fSize, _ := strconv.Atoi(ui.TblFiles.GetCell(idx, 5).Text)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell(" ✓ "))
			ui.TblFiles.GetCell(idx, 0).SetTextColor(tcell.ColorRed)
			ui.TblFiles.GetCell(idx, 1).SetTextColor(tcell.ColorRed)
			sel = append(sel, selecao{fName: fName, fSize: int64(fSize), fType: "FILE"})
			displaySelection()
		} else {
			// UNSELECT FILE
			fName := filepath.Join(fm.Cwd, ui.TblFiles.GetCell(idx, 1).Text)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell("   "))
			ui.TblFiles.GetCell(idx, 0).SetTextColor(tcell.ColorYellow)
			ui.TblFiles.GetCell(idx, 1).SetTextColor(tcell.ColorYellow)
			sel = findAndDelete(sel, selecao{fName: fName, fSize: 0, fType: "FILE"})
			displaySelection()
		}
	} else {
		if ui.TblFiles.GetCell(idx, 0).Text == "   " {
			// SELECT FOLDER
			fName := filepath.Join(fm.Cwd, ui.TblFiles.GetCell(idx, 1).Text)
			fSize, _ := utils.DirSize(fName)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell(" ✓ "))
			ui.TblFiles.GetCell(idx, 0).SetTextColor(tcell.ColorRed)
			ui.TblFiles.GetCell(idx, 1).SetTextColor(tcell.ColorRed)
			sel = append(sel, selecao{fName: fName, fSize: fSize, fType: "FOLDER"})
			displaySelection()
		} else {
			// UNSELECT FOLDER
			fName := filepath.Join(fm.Cwd, ui.TblFiles.GetCell(idx, 1).Text)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell("   "))
			ui.TblFiles.GetCell(idx, 0).SetTextColor(tcell.ColorYellow)
			ui.TblFiles.GetCell(idx, 1).SetTextColor(tcell.ColorYellow)
			sel = findAndDelete(sel, selecao{fName: fName, fSize: 0, fType: "FOLDER"})
			displaySelection()
		}
	}
	// Move cursor to next line
	if idx < ui.TblFiles.GetRowCount()-1 {
		ui.TblFiles.Select(idx+1, 0)
	}
}

// ****************************************************************************
// displaySelection()
// ****************************************************************************
func displaySelection() {
	nFiles := 0
	nFolders := 0
	nSize := 0
	for _, s := range sel {
		nSize += int(s.fSize)
		if s.fType == "FILE" {
			nFiles++
		} else {
			nFolders++
		}
	}
	infos := map[string]string{
		"0Files":   fmt.Sprintf("%d", nFiles),
		"1Folders": fmt.Sprintf("%d", nFolders),
		"2Size":    fmt.Sprintf("%d Bytes (%s)", nSize, utils.HumanFileSize(float64(nSize))),
	}
	ui.DisplayMap(ui.TxtSelection, infos)
}

// ****************************************************************************
// findAndDelete()
// ****************************************************************************
func findAndDelete(s []selecao, item selecao) []selecao {
	index := 0
	for _, i := range s {
		if i.fName != item.fName {
			s[index] = i
			index++
		}
	}
	return s[:index]
}

// ****************************************************************************
// welcome()
// ****************************************************************************
func welcome() (string, string) {
	w1 := ":: Welcome to " + conf.APP_STRING + " :"
	w2 := conf.APP_NAME + " version " + conf.APP_VERSION + " - " + conf.APP_URL + "\n"
	hst, err := os.Hostname()
	if err != nil {
		hst = "localhost"
	}
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
	ui.LblHostname.SetText(hst)
	return w1, w2
}

/*
F1	Aide
F2	Menu
F3	Voir
F4	Modif
F5	Copier
F6	RenDep
F7	CréRep
F8	Suppr
F9	MenuDér
F10	Quitter

ps -eo pid,user,lstart,cmd
ps -p <pid> -o lstart
kill -STOP <pid> => pause
kill -CONT <pid> => reprend

Create Folder
Create File from ~/Modèles
Create Link
Zip file/folder
Delete file/folder
Move file/folder
Rename file/folder
Hash file
Encrypt file
Sort by names/dates/size
List hidden files

Edit file
Hexedit file

List process
Kill process
Pause process
Cont process
Nice process

*/
