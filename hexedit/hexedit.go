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
package hexedit

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"bufio"
	"errors"
	"fmt"
	"gosh/conf"
	"gosh/dialog"
	"gosh/preview"
	"gosh/ui"
	"gosh/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ****************************************************************************
// VARS
// ****************************************************************************
var (
	DlgOpen               *dialog.Dialog
	headerBackgroundColor = tcell.ColorBlue
	headerTextColor       = tcell.ColorWhite
	CurrentHexFile        string
)

// ****************************************************************************
// OpenFile()
// ****************************************************************************
func OpenFile(fName string) {
	CurrentHexFile = fName
	f, err := os.Open(fName)
	if err != nil {
		ui.SetStatus(err.Error())
	}
	defer f.Close()

	offset := 0
	r, c := 0, 0
	br := bufio.NewReader(f)
	var ascii string
	ui.TblHexEdit.Clear()

	ui.TblHexEdit.SetCell(0, 0, tview.NewTableCell("offset").SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor).SetAlign(tview.AlignCenter))
	for i := 0; i < 16; i++ {
		ui.TblHexEdit.SetCell(0, i+1, tview.NewTableCell(fmt.Sprintf("%02X", i)).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	}
	for i := 0; i < 16; i++ {
		ui.TblHexEdit.SetCell(0, i+17, tview.NewTableCell(fmt.Sprintf("%01X", i)).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	}

	for {
		b, err := br.ReadByte()

		if err != nil && !errors.Is(err, io.EOF) {
			ui.SetStatus(err.Error())
			break
		}
		c = offset % 16
		// Display the address in first column
		if c == 0 {
			ui.TblHexEdit.SetCell(r+1, c, tview.NewTableCell(fmt.Sprintf(" %08X", offset)).SetTextColor(tcell.ColorBlue))
		}

		// Display the Hex value
		ui.TblHexEdit.SetCell(r+1, c+1, tview.NewTableCell(fmt.Sprintf("%02X", b)))
		// Add the ascii value to the ascii string
		if isASCII(string(b)) {
			ascii = ascii + string(b)
		} else {
			ascii = ascii + "."
		}
		/*
			if unicode.IsPrint(rune(b)) {
				ascii = ascii + string(b)
			} else {
				ascii = ascii + "."
			}
		*/

		// Display the ascii string
		if c == 15 {
			for i := 0; i < len(ascii); i++ {
				ui.TblHexEdit.SetCell(r+1, c+2+i, tview.NewTableCell(string(ascii[i])).SetTextColor(tcell.ColorYellow))
			}
			ascii = ""
			r++
		}
		offset++
		if err != nil { // end of file
			break
		}
	}
	offset--
	ui.TxtHexName.SetText(fmt.Sprintf("[white]File [yellow]%s[white] (Size [yellow]%d[white] bytes, [yellow]%s[white])", fName, offset, utils.HumanFileSize(float64(offset))))
	ui.TblHexEdit.SetFixed(1, 0)
	ui.TblHexEdit.Select(1, 0)
	ui.TblHexEdit.ScrollToBeginning()
	preview.DisplayExif(fName)
	ui.PgsApp.SwitchToPage(ui.GetCurrentScreen())
	ui.App.SetFocus(ui.TxtPrompt)
	ui.SetStatus(fmt.Sprintf("File %s successfully open", fName))
}

// ****************************************************************************
// DoOpen()
// ****************************************************************************
func DoOpen(path string) {
	DlgOpen = DlgOpen.Input("Open binary file", // Title
		"Please, enter the name of the binary file to open :", // Message
		path,
		confirmOpen,
		0,
		"hexedit", ui.TblHexEdit) // Focus return
	ui.PgsApp.AddPage("dlgOpenHex", DlgOpen.Popup(), true, false)
	ui.PgsApp.ShowPage("dlgOpenHex")
}

// ****************************************************************************
// confirmOpen()
// ****************************************************************************
func confirmOpen(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_OK {
		fName := DlgOpen.Value
		OpenFile(fName)
	}
}

// ****************************************************************************
// Close()
// ****************************************************************************
func Close() {
	// TODO : Save any modification on file
	ui.TblHexEdit.Clear()
	ui.TxtHexName.Clear()
	ui.TxtFileInfo.Clear()
	CurrentHexFile = ""
	ui.SetStatus("File closed")
}

// ****************************************************************************
// SelfInit()
// ****************************************************************************
func SelfInit(a any) {
	if ui.CurrentMode == ui.ModeFiles {
		if CurrentHexFile == "" {
			idx, _ := ui.TblFiles.GetSelection()
			fName := filepath.Join(conf.Cwd, strings.TrimSpace(ui.TblFiles.GetCell(idx, 2).Text))
			ui.AddNewScreen(ui.ModeHexEdit, nil, nil)
			OpenFile(fName)
			ui.App.SetFocus(ui.TxtPrompt)
		} else {
			ui.AddNewScreen(ui.ModeHexEdit, nil, nil)
			ui.App.SetFocus(ui.TxtPrompt)
		}
	} else {
		ui.AddNewScreen(ui.ModeHexEdit, nil, nil)
		ui.App.SetFocus(ui.TxtPrompt)
	}
}

// ****************************************************************************
// isASCII()
// ****************************************************************************
func isASCII(s string) bool {
	for _, c := range s {
		if c < 32 || c > unicode.MaxASCII {
			return false
		}
	}
	return true
}
