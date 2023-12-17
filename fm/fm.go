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

// ****************************************************************************
// fm is the File Manager module
// ****************************************************************************

import (
	"fmt"
	"gosh/conf"
	"gosh/dialog"
	"gosh/menu"
	"gosh/preview"
	"gosh/ui"
	"gosh/utils"
	"io/fs"
	"os"
	"sort"
	"strconv"

	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type SortColumn int

type selecao struct {
	fName string
	fSize int64
	fType string
}

const (
	SORT_NAME SortColumn = iota
	SORT_TIME
	SORT_SIZE
)

const (
	SORT_ASCENDING  = 0
	SORT_DESCENDING = 1
)

// ****************************************************************************
// GLOBALS
// ****************************************************************************
var sortColumn = SORT_NAME
var sortOrder = SORT_ASCENDING
var (
	Cwd          string
	Hidden       bool
	MnuFiles     *menu.Menu
	MnuFilesSort *menu.Menu
	DlgConfirm   *dialog.Dialog
	sel          []selecao
)

// ****************************************************************************
// SetFilesMenu()
// ****************************************************************************
func SetFilesMenu() {
	MnuFiles = MnuFiles.New("Actions", "files", ui.TblFiles)
	MnuFiles.AddItem("mnuEdit", "Edit", DoDelete, true)
	MnuFiles.AddItem("mnuOpen", "Open", DoDelete, true)
	MnuFiles.AddItem("mnuDelete", "Delete", DoDelete, true)
	MnuFiles.AddItem("mnuRename", "Rename", DoRename, true)
	MnuFiles.AddItem("mnuMove", "Move", DoRename, true)
	MnuFiles.AddItem("mnuCreateFile", "Create File", DoCreateFile, true)
	MnuFiles.AddItem("mnuCreateFolder", "Create Folder", DoCreateFolder, true)
	MnuFiles.AddItem("mnuCreateLink", "Create Link", DoCreateLink, true)
	MnuFiles.AddItem("mnuZip", "Zip", DoDelete, true)
	MnuFiles.AddItem("mnuHashes", "Get Hashes", DoDelete, true)
	MnuFiles.AddItem("mnuEncrypt", "Encrypt", DoDelete, true)
	MnuFiles.AddItem("mnuShowHiddenFiles", "Show hidden files", DoSwitchHiddenFiles, true)
	ui.PgsApp.AddPage("dlgFileAction", MnuFiles.Popup(), true, false)

	MnuFilesSort = MnuFilesSort.New("Sort by", "files", ui.TblFiles)
	MnuFilesSort.AddItem("mnuSortNameA", "Name Ascending", doSortNameA, false)
	MnuFilesSort.AddItem("mnuSortNameD", "Name Descending", doSortNameD, true)
	MnuFilesSort.AddItem("mnuSortSizeA", "Size Ascending", doSortSizeA, true)
	MnuFilesSort.AddItem("mnuSortSizeD", "Size Descending", doSortSizeD, true)
	MnuFilesSort.AddItem("mnuSortTimeA", "Time Ascending", doSortTimeA, true)
	MnuFilesSort.AddItem("mnuSortTimeD", "Time Descending", doSortTimeD, true)
	ui.PgsApp.AddPage("dlgFileSort", MnuFilesSort.Popup(), true, false)

}

// ****************************************************************************
// ShowMenu()
// ****************************************************************************
func ShowMenu() {
	idx, _ := ui.TblFiles.GetSelection()
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 3).Text)
	// fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 1).Text)
	if targetType == "FOLDER" {
		MnuFiles.SetEnabled("mnuEdit", false)
		MnuFiles.SetEnabled("mnuOpen", false)
		MnuFiles.SetEnabled("mnuEncrypt", false)
	}
	if targetType == "FILE" {
		MnuFiles.SetEnabled("mnuEdit", true)
		MnuFiles.SetEnabled("mnuOpen", true)
		MnuFiles.SetEnabled("mnuEncrypt", true)
	}
	if Hidden {
		MnuFiles.SetLabel("mnuShowHiddenFiles", "Hide hidden files")
	} else {
		MnuFiles.SetLabel("mnuShowHiddenFiles", "Show hidden files")
	}
	ui.PgsApp.ShowPage("dlgFileAction")
}

// ****************************************************************************
// ShowMenuSort()
// ****************************************************************************
func ShowMenuSort() {
	ui.PgsApp.ShowPage("dlgFileSort")
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
		// DlgConfirm = DlgConfirm.New("Delete File", dialog.YesNo, "Are you sure ?", []string{"Yes", "No"}, ProcessDelete, "files", ui.TblFiles)
		// ui.PgsApp.AddPage("dlgFileAction", DlgConfirm.Popup(), true, false)
	} else {
		ui.SetStatus("Deleting folder " + fName)
	}
}

/*
func ProcessDelete(bIndex int, bLabel string) {
	ui.PgsApp.SwitchToPage(DlgConfirm.Parent)
	ui.App.SetFocus(DlgConfirm.FFocus)
}
*/

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
	ShowFiles()
}

// ****************************************************************************
// ShowFiles()
// ****************************************************************************
func ShowFiles() {
	// ui.TxtSelection.Clear()
	ui.PgsApp.SwitchToPage("files")
	files, err := os.ReadDir(Cwd)
	if err != nil {
		ui.SetStatus(err.Error())
	}

	ui.TblFiles.Clear()
	ui.TxtFileInfo.Clear()
	iStart := 0
	iFile := 0
	if Cwd != "/" {
		ui.TblFiles.SetCell(0, 0, tview.NewTableCell("   "))
		ui.TblFiles.SetCell(0, 1, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 2, tview.NewTableCell("..").SetTextColor(tcell.ColorYellow))
		ui.TblFiles.SetCell(0, 3, tview.NewTableCell("<UP>"))
		ui.TblFiles.SetCell(0, 4, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 5, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 6, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(0, 7, tview.NewTableCell(" "))
		iStart = 1
	}
	ui.TxtPath.SetText(Cwd)
	switch sortColumn {
	case SORT_NAME:
		if sortOrder == SORT_ASCENDING {
			SortFileNameAscend(files)
		} else {
			SortFileNameDescend(files)
		}
	case SORT_SIZE:
		if sortOrder == SORT_ASCENDING {
			SortFileSizeAscend(files)
		} else {
			SortFileSizeDescend(files)
		}
	case SORT_TIME:
		if sortOrder == SORT_ASCENDING {
			SortFileModAscend(files)
		} else {
			SortFileModDescend(files)
		}
	}
	for _, file := range files {
		if !Hidden && file.Name()[0] == '.' { // Don't want to see hidden files ?
			continue
		}
		ui.TblFiles.SetCell(iFile+iStart, 0, tview.NewTableCell("   "))
		ui.TblFiles.SetCell(iFile+iStart, 1, tview.NewTableCell(" "))
		ui.TblFiles.SetCell(iFile+iStart, 2, tview.NewTableCell(file.Name()).SetTextColor(tcell.ColorYellow))
		fi, err := file.Info()
		if err == nil {
			ui.TblFiles.SetCell(iFile+iStart, 3, tview.NewTableCell(fi.ModTime().String()[0:19]))
			if fi.IsDir() {
				ui.TblFiles.SetCell(iFile+iStart, 4, tview.NewTableCell("  FOLDER"))
				ui.TblFiles.SetCell(0, 7, tview.NewTableCell(" "))
				ui.TblFiles.GetCell(iFile+iStart, 2).SetTextColor(tcell.ColorLightGreen)
			} else {
				if fi.Mode().String()[0] == 'L' {
					ui.TblFiles.SetCell(iFile+iStart, 1, tview.NewTableCell("🔗"))
					ui.TblFiles.SetCell(iFile+iStart, 4, tview.NewTableCell("  LINK"))

					lnk, err := os.Readlink(filepath.Join(Cwd, ui.TblFiles.GetCell(iFile+iStart, 2).Text))
					if err == nil {
						ui.TblFiles.SetCell(iFile+iStart, 7, tview.NewTableCell(lnk))
					} else {
						ui.TblFiles.SetCell(iFile+iStart, 7, tview.NewTableCell(err.Error()))
					}
				} else {
					ui.TblFiles.SetCell(iFile+iStart, 4, tview.NewTableCell("  FILE"))
					ui.TblFiles.SetCell(0, 7, tview.NewTableCell(" "))
					// Is the file executable ?
					if fi.Mode()&0111 != 0 {
						ui.TblFiles.SetCell(iFile+iStart, 1, tview.NewTableCell("⚙"))
						ui.TblFiles.GetCell(iFile+iStart, 2).SetTextColor(tcell.ColorLightYellow)
					}
				}
			}
			ui.TblFiles.SetCell(iFile+iStart, 5, tview.NewTableCell(fi.Mode().String()))
			ui.TblFiles.SetCell(iFile+iStart, 6, tview.NewTableCell(strconv.FormatInt(fi.Size(), 10)).SetAlign(tview.AlignRight))
		}
		iFile++
	}
	ui.TblFiles.Select(0, 0)
}

// ****************************************************************************
// RefreshMe()
// ****************************************************************************
func RefreshMe() {
	ShowFiles()
	ui.App.SetFocus(ui.TblFiles)
}

// ****************************************************************************
// SortFileNameAscend()
// ****************************************************************************
func SortFileNameAscend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		return strings.ToUpper(files[i].Name()) < strings.ToUpper(files[j].Name())
	})
}

// ****************************************************************************
// SortFileNameDescend()
// ****************************************************************************
func SortFileNameDescend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		return strings.ToUpper(files[i].Name()) > strings.ToUpper(files[j].Name())
	})
}

// ****************************************************************************
// SortFileSizeAscend()
// ****************************************************************************
func SortFileSizeAscend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		ifs, _ := files[i].Info()
		jfs, _ := files[j].Info()
		return ifs.Size() < jfs.Size()
	})
}

// ****************************************************************************
// SortFileSizeDescend()
// ****************************************************************************
func SortFileSizeDescend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		ifs, _ := files[i].Info()
		jfs, _ := files[j].Info()
		return ifs.Size() > jfs.Size()
	})
}

// ****************************************************************************
// SortFileModAscend()
// ****************************************************************************
func SortFileModAscend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		ifs, _ := files[i].Info()
		jfs, _ := files[j].Info()
		return ifs.ModTime().Unix() < jfs.ModTime().Unix()
	})
}

// ****************************************************************************
// SortFileModDescend()
// ****************************************************************************
func SortFileModDescend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		ifs, _ := files[i].Info()
		jfs, _ := files[j].Info()
		return ifs.ModTime().Unix() > jfs.ModTime().Unix()
	})
}

// ****************************************************************************
// doSortNameA()
// ****************************************************************************
func doSortNameA() {
	sortColumn = SORT_NAME
	sortOrder = SORT_ASCENDING
	MnuFilesSort.SetEnabled("mnuSortNameA", false)
	MnuFilesSort.SetEnabled("mnuSortNameD", true)
	MnuFilesSort.SetEnabled("mnuSortSizeA", true)
	MnuFilesSort.SetEnabled("mnuSortSizeD", true)
	MnuFilesSort.SetEnabled("mnuSortTimeA", true)
	MnuFilesSort.SetEnabled("mnuSortTimeD", true)
	ShowFiles()
}

// ****************************************************************************
// doSortNameD()
// ****************************************************************************
func doSortNameD() {
	sortColumn = SORT_NAME
	sortOrder = SORT_DESCENDING
	MnuFilesSort.SetEnabled("mnuSortNameA", true)
	MnuFilesSort.SetEnabled("mnuSortNameD", false)
	MnuFilesSort.SetEnabled("mnuSortSizeA", true)
	MnuFilesSort.SetEnabled("mnuSortSizeD", true)
	MnuFilesSort.SetEnabled("mnuSortTimeA", true)
	MnuFilesSort.SetEnabled("mnuSortTimeD", true)
	ShowFiles()
}

// ****************************************************************************
// doSortSizeA()
// ****************************************************************************
func doSortSizeA() {
	sortColumn = SORT_SIZE
	sortOrder = SORT_ASCENDING
	MnuFilesSort.SetEnabled("mnuSortNameA", true)
	MnuFilesSort.SetEnabled("mnuSortNameD", true)
	MnuFilesSort.SetEnabled("mnuSortSizeA", false)
	MnuFilesSort.SetEnabled("mnuSortSizeD", true)
	MnuFilesSort.SetEnabled("mnuSortTimeA", true)
	MnuFilesSort.SetEnabled("mnuSortTimeD", true)
	ShowFiles()
}

// ****************************************************************************
// doSortSizeD()
// ****************************************************************************
func doSortSizeD() {
	sortColumn = SORT_SIZE
	sortOrder = SORT_DESCENDING
	MnuFilesSort.SetEnabled("mnuSortNameA", true)
	MnuFilesSort.SetEnabled("mnuSortNameD", true)
	MnuFilesSort.SetEnabled("mnuSortSizeA", true)
	MnuFilesSort.SetEnabled("mnuSortSizeD", false)
	MnuFilesSort.SetEnabled("mnuSortTimeA", true)
	MnuFilesSort.SetEnabled("mnuSortTimeD", true)
	ShowFiles()
}

// ****************************************************************************
// doSortTimeA()
// ****************************************************************************
func doSortTimeA() {
	sortColumn = SORT_TIME
	sortOrder = SORT_ASCENDING
	MnuFilesSort.SetEnabled("mnuSortNameA", true)
	MnuFilesSort.SetEnabled("mnuSortNameD", true)
	MnuFilesSort.SetEnabled("mnuSortSizeA", true)
	MnuFilesSort.SetEnabled("mnuSortSizeD", true)
	MnuFilesSort.SetEnabled("mnuSortTimeA", false)
	MnuFilesSort.SetEnabled("mnuSortTimeD", true)
	ShowFiles()
}

// ****************************************************************************
// doSortTimeD()
// ****************************************************************************
func doSortTimeD() {
	sortColumn = SORT_TIME
	sortOrder = SORT_DESCENDING
	MnuFilesSort.SetEnabled("mnuSortNameA", true)
	MnuFilesSort.SetEnabled("mnuSortNameD", true)
	MnuFilesSort.SetEnabled("mnuSortSizeA", true)
	MnuFilesSort.SetEnabled("mnuSortSizeD", true)
	MnuFilesSort.SetEnabled("mnuSortTimeA", true)
	MnuFilesSort.SetEnabled("mnuSortTimeD", false)
	ShowFiles()
}

// ****************************************************************************
// ProceedFileAction()
// ****************************************************************************
func ProceedFileAction() {
	idx, _ := ui.TblFiles.GetSelection()
	// TODO : manage LINK
	targetType := strings.TrimSpace(ui.TblFiles.GetCell(idx, 4).Text)
	if targetType == "LINK" {
		targetType = "FILE"
		fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 2).Text)
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
		fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 2).Text)
		ui.FrmFileInfo.Clear()

		mtype, xmtype := preview.DisplayFilePreview(fName)
		size, _ := strconv.ParseFloat(ui.TblFiles.GetCell(idx, 6).Text, 64)
		if size <= conf.HASH_THRESHOLD_SIZE {
			infos := map[string]string{
				"00Name":          ui.TblFiles.GetCell(idx, 2).Text,
				"01Change Date":   ui.TblFiles.GetCell(idx, 3).Text,
				"02Access":        ui.TblFiles.GetCell(idx, 5).Text,
				"03Size":          ui.TblFiles.GetCell(idx, 6).Text + " Bytes (" + utils.HumanFileSize(size) + ")",
				"04Mime Type":     mtype,
				"05Extended Mime": xmtype,
			}
			ui.DisplayMap(ui.FrmFileInfo, infos)

		} else {
			infos := map[string]string{
				"00Name":        ui.TblFiles.GetCell(idx, 2).Text,
				"01Change Date": ui.TblFiles.GetCell(idx, 3).Text,
				"02Access":      ui.TblFiles.GetCell(idx, 5).Text,
				"03Size":        ui.TblFiles.GetCell(idx, 6).Text + " Bytes (" + utils.HumanFileSize(size) + ")",
			}
			ui.DisplayMap(ui.FrmFileInfo, infos)
			ui.TxtFileInfo.SetText("VERY BIG FILE, can't display a preview.")
		}
	} else { //  or type(readlink)==folder
		Cwd = filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 2).Text)
		ui.FrmFileInfo.Clear()
		nFiles, nFolders, err := utils.NumberOfFilesAndFolders(Cwd)
		if err != nil {
			ui.SetStatus(err.Error())
		}
		infos := map[string]string{
			"00Name":    Cwd,
			"01Files":   strconv.Itoa(nFiles),
			"02Folders": strconv.Itoa(nFolders),
		}
		ui.DisplayMap(ui.FrmFileInfo, infos)
		ShowFiles()
		ui.App.SetFocus(ui.TblFiles)
	}
	ui.App.Sync()
}

// ****************************************************************************
// ProceedFileSelect()
// ****************************************************************************
func ProceedFileSelect() {
	idx, _ := ui.TblFiles.GetSelection()
	if strings.TrimSpace(ui.TblFiles.GetCell(idx, 4).Text) == "FILE" {
		if ui.TblFiles.GetCell(idx, 0).Text == "   " {
			// SELECT FILE
			fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 2).Text)
			fSize, _ := strconv.Atoi(ui.TblFiles.GetCell(idx, 6).Text)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell(" ✓ "))
			ui.TblFiles.GetCell(idx, 0).SetTextColor(conf.COLOR_SELECTED)
			ui.TblFiles.GetCell(idx, 1).SetTextColor(conf.COLOR_SELECTED)
			ui.TblFiles.GetCell(idx, 2).SetTextColor(conf.COLOR_SELECTED)
			sel = append(sel, selecao{fName: fName, fSize: int64(fSize), fType: "FILE"})
			displaySelection()
		} else {
			// UNSELECT FILE
			fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 2).Text)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell("   "))
			if ui.TblFiles.GetCell(idx, 1).Text == "⚙" {
				ui.TblFiles.GetCell(idx, 0).SetTextColor(conf.COLOR_EXECUTABLE)
				ui.TblFiles.GetCell(idx, 1).SetTextColor(conf.COLOR_EXECUTABLE)
				ui.TblFiles.GetCell(idx, 2).SetTextColor(conf.COLOR_EXECUTABLE)
			} else {
				ui.TblFiles.GetCell(idx, 0).SetTextColor(conf.COLOR_FILE)
				ui.TblFiles.GetCell(idx, 1).SetTextColor(conf.COLOR_FILE)
				ui.TblFiles.GetCell(idx, 2).SetTextColor(conf.COLOR_FILE)
			}
			sel = findAndDelete(sel, selecao{fName: fName, fSize: 0, fType: "FILE"})
			displaySelection()
		}
	} else {
		if ui.TblFiles.GetCell(idx, 0).Text == "   " {
			// SELECT FOLDER
			fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 2).Text)
			fSize, _ := utils.DirSize(fName)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell(" ✓ "))
			ui.TblFiles.GetCell(idx, 0).SetTextColor(conf.COLOR_SELECTED)
			ui.TblFiles.GetCell(idx, 1).SetTextColor(conf.COLOR_SELECTED)
			ui.TblFiles.GetCell(idx, 2).SetTextColor(conf.COLOR_SELECTED)
			sel = append(sel, selecao{fName: fName, fSize: fSize, fType: "FOLDER"})
			displaySelection()
		} else {
			// UNSELECT FOLDER
			fName := filepath.Join(Cwd, ui.TblFiles.GetCell(idx, 2).Text)
			ui.SetStatus(fName)
			ui.TblFiles.SetCell(idx, 0, tview.NewTableCell("   "))
			ui.TblFiles.GetCell(idx, 0).SetTextColor(conf.COLOR_FOLDER)
			ui.TblFiles.GetCell(idx, 1).SetTextColor(conf.COLOR_FOLDER)
			ui.TblFiles.GetCell(idx, 2).SetTextColor(conf.COLOR_FOLDER)
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
		"00Files":   fmt.Sprintf("%d", nFiles),
		"01Folders": fmt.Sprintf("%d", nFolders),
		"02Size":    fmt.Sprintf("%d Bytes (%s)", nSize, utils.HumanFileSize(float64(nSize))),
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
