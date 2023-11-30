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
package fm

import (
	"gosh/menu"
	"gosh/ui"
	"os"
	"strconv"

	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	Cwd      string
	Hidden   bool
	MnuFiles *menu.Menu
)

// ****************************************************************************
// SetFilesMenu()
// ****************************************************************************
func SetFilesMenu() {
	MnuFiles = MnuFiles.New("Actions", "files", ui.TblFiles)

	MnuFiles.AddItem("mnuDelete", "Delete", DoDelete, true)
	MnuFiles.AddItem("mnuRename", "Rename", DoRename, true)
	MnuFiles.AddItem("mnuCreateFile", "Create File", DoCreateFile, true)
	MnuFiles.AddItem("mnuCreateFolder", "Create Folder", DoCreateFolder, true)
	MnuFiles.AddItem("mnuCreateLink", "Create Link", DoCreateLink, true)
	MnuFiles.AddItem("mnuZip", "Zip", DoDelete, true)
	MnuFiles.AddItem("mnuHashes", "Get Hashes", DoDelete, true)
	MnuFiles.AddItem("mnuEncrypt", "Encrypt", DoDelete, true)
	MnuFiles.AddItem("mnuShowHiddenFiles", "Show hidden files", DoSwitchHiddenFiles, true)
	MnuFiles.AddItem("mnuSort", "Sort by names", DoDelete, true)

	ui.PgsApp.AddPage("dlgFileAction", MnuFiles.Popup(), true, false)
}

// ****************************************************************************
// InitMenu()
// ****************************************************************************
func ShowMenu() {
	idx, _ := ui.TblFiles.GetSelection()
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	// fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 1).Text)
	if targetType == "FILE" {
		MnuFiles.SetEnabled("mnuZip", false)
	}
	if Hidden {
		MnuFiles.SetLabel("mnuShowHiddenFiles", "Hide hidden files")
	} else {
		MnuFiles.SetLabel("mnuShowHiddenFiles", "Show hidden files")
	}
	ui.PgsApp.ShowPage("dlgFileAction")
}

// ****************************************************************************
// DoDelete()
// ****************************************************************************
func DoDelete() {
	idx, _ := ui.TblFiles.GetSelection()
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 1).Text)
	if targetType == "FILE" {
		ui.SetStatus("Deleting file " + fName)
	} else {
		ui.SetStatus("Deleting folder " + fName)
	}
}

// ****************************************************************************
// DoRename()
// ****************************************************************************
func DoRename() {
	idx, _ := ui.TblFiles.GetSelection()
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 1).Text)
	if targetType == "FILE" {
		ui.SetStatus("Renaming file " + fName)
	} else {
		ui.SetStatus("Renaming folder " + fName)
	}
}

// ****************************************************************************
// DoCreateFile()
// ****************************************************************************
func DoCreateFile() {
	idx, _ := ui.TblFiles.GetSelection()
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 1).Text)
	if targetType == "FILE" {
		ui.SetStatus("Renaming file " + fName)
	} else {
		ui.SetStatus("Renaming folder " + fName)
	}
}

// ****************************************************************************
// DoCreateFolder()
// ****************************************************************************
func DoCreateFolder() {
	idx, _ := ui.TblFiles.GetSelection()
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 1).Text)
	if targetType == "FILE" {
		ui.SetStatus("Renaming file " + fName)
	} else {
		ui.SetStatus("Renaming folder " + fName)
	}
}

// ****************************************************************************
// DoCreateLink()
// ****************************************************************************
func DoCreateLink() {
	idx, _ := ui.TblFiles.GetSelection()
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 1).Text)
	if targetType == "FILE" {
		ui.SetStatus("Creating link for file " + fName)
	} else {
		ui.SetStatus("Creating link for folder " + fName)
	}
}

// ****************************************************************************
// DoSwitchHiddenFiles()
// ****************************************************************************
func DoSwitchHiddenFiles() {
	Hidden = !Hidden
}

// ****************************************************************************
// ShowFiles()
// ****************************************************************************
func ShowFiles() {
	ui.TxtSelection.Clear()
	ui.PgsApp.SwitchToPage("files")
	files, err := os.ReadDir(Cwd)
	if err != nil {
		ui.SetStatus(err.Error())
	}

	ui.TblFiles.Clear()
	ui.TxtFileInfo.Clear()
	iStart := 0
	if Cwd != "/" {
		ui.TblFiles.SetCell(0, 0, tview.NewTableCell("   "))
		ui.TblFiles.SetCell(0, 1, tview.NewTableCell("..").SetTextColor(tcell.ColorYellow))
		ui.TblFiles.SetCell(0, 2, tview.NewTableCell("<UP>"))
		ui.TblFiles.SetCell(0, 3, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 4, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 5, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 6, tview.NewTableCell(" "))
		iStart = 1
	}
	ui.TxtPath.SetText(Cwd)
	for i, file := range files {
		if !Hidden && file.Name()[0] == '.' {
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
					lnk, err := os.Readlink(filepath.Join(Cwd, ui.TblFiles.GetCell(i+iStart, 1).Text))
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
}

/*
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
*/
