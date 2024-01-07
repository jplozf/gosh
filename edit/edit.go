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
package edit

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
	"gosh/conf"
	"gosh/ui"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type editfile struct {
	buffer *femto.Buffer
	fName  string
}

// ****************************************************************************
// GLOBALS
// ****************************************************************************
var (
	openFiles   []editfile
	currentFile editfile
)

// ****************************************************************************
// SwitchToEditor()
// ****************************************************************************
func SwitchToEditor() {
	ui.CurrentMode = ui.ModeTextEdit
	ui.SetTitle("Editor")
	ui.LblKeys.SetText("F1=Help F2=Shell F3=Files F4=Process F12=Exit\nCtrl+S=Save")
	ui.PgsApp.SwitchToPage("editor")
	ui.App.SetFocus(ui.EdtMain)
}

// ****************************************************************************
// OpenFile()
// ****************************************************************************
func OpenFile(fName string) {
	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	ui.EdtMain.SetRuntimeFiles(runtime.Files)
	content, err := ioutil.ReadFile(fName)
	if err != nil {
		ui.SetStatus(fmt.Sprintf("Could not read %v: %v", fName, err))
	} else {
		currentFile.fName = fName
		currentFile.buffer = femto.NewBufferFromString(string(content), currentFile.fName)
		ui.EdtMain.OpenBuffer(currentFile.buffer)
		ui.EdtMain.SetColorscheme(colorscheme)
		ui.EdtMain.SetTitleAlign(tview.AlignRight)
		openFiles = append(openFiles, currentFile)
		go UpdateStatus()
	}
}

// ****************************************************************************
// SaveFile()
// ****************************************************************************
func SaveFile() {
	err := ioutil.WriteFile(currentFile.fName, []byte(currentFile.buffer.String()), 0600)
	if err == nil {
		ui.SetStatus(fmt.Sprintf("File %s successfully saved", currentFile.fName))
		currentFile.buffer.IsModified = false
	} else {
		ui.SetStatus(err.Error())
	}
}

// ****************************************************************************
// UpdateStatus()
// ****************************************************************************
func UpdateStatus() {
	var status string
	for {
		time.Sleep(100 * time.Millisecond)
		ui.App.QueueUpdateDraw(func() {
			ui.TxtEditName.SetText(currentFile.fName)
			if currentFile.buffer.Modified() {
				status = conf.ICON_MODIFIED
			} else {
				status = " "
			}
			x := currentFile.buffer.Cursor.X + 1
			y := currentFile.buffer.Cursor.Y + 1
			ui.EdtMain.SetTitle(fmt.Sprintf("[ Ln %d, Col %d %s ]", y, x, status))
			for i, f := range openFiles {
				if f.buffer.Modified() {
					ui.TblOpenFiles.SetCell(i, 0, tview.NewTableCell(conf.ICON_MODIFIED))
				} else {
					ui.TblOpenFiles.SetCell(i, 0, tview.NewTableCell(" "))
				}
				ui.TblOpenFiles.SetCell(i, 1, tview.NewTableCell(filepath.Base(f.fName)))
			}
		})
	}
}
