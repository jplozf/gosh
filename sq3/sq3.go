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
package sq3

import (
	"bufio"
	"database/sql"
	"fmt"
	"gosh/conf"
	"gosh/dialog"
	"gosh/edit"
	"gosh/menu"
	"gosh/ui"
	"gosh/utils"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gdamore/tcell/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rivo/tview"
)

// ****************************************************************************
// sq3 is the SQLite3 interface
// ****************************************************************************

var (
	CurrentDatabaseName string
	CurrentDB           *sql.DB
	DlgOpenDB           *dialog.Dialog
	DlgCloseDB          *dialog.Dialog
	ACmd                []string
	ICmd                int
	root                *tview.TreeNode
	MnuSQL              *menu.Menu
)
var headerBackgroundColor = tcell.ColorDarkGreen
var headerTextColor = tcell.ColorYellow

// ****************************************************************************
// Xeq()
// ****************************************************************************
func Xeq(c string) {
	c = strings.TrimSpace(c)
	/*
		if c[0] == '!' {
			cmd.Xeq(c)
		} else
	*/
	if c[0] == '.' {
		ACmd = append(ACmd, c)
		if strings.HasPrefix(strings.ToUpper(c), ".TABLE") {
			doSelect("SELECT name FROM sqlite_master WHERE type ='table' AND name NOT LIKE 'sqlite_%';")
		} else {
			if strings.HasPrefix(strings.ToUpper(c), ".DATABASE") {
				doSelect("PRAGMA database_list;")
			} else {
				if strings.HasPrefix(strings.ToUpper(c), ".SCHEMA") {
					tokens := strings.Fields(c)
					if len(tokens) > 1 {
						table := tokens[1]
						doSelect(fmt.Sprintf("SELECT sql FROM sqlite_master WHERE name = '%s';", table))
					} else {
						ui.SetStatus(tview.Escape("Too few arguments for .SCHEMA [table]"))
					}
				} else {
					if strings.HasPrefix(strings.ToUpper(c), ".COLUMNS") {
						tokens := strings.Fields(c)
						if len(tokens) > 1 {
							table := tokens[1]
							doSelect(fmt.Sprintf("PRAGMA table_info(%s);", table))
						} else {
							ui.SetStatus(tview.Escape("Too few arguments for .COLUMNS [table]"))
						}
					} else {
						if strings.HasPrefix(strings.ToUpper(c), ".OPEN") {
							tokens := strings.Fields(c)
							if len(tokens) > 1 {
								db := tokens[1]
								OpenDB(db)
							} else {
								ui.SetStatus(tview.Escape("Too few arguments for .OPEN [database]"))
							}
						} else {
							if strings.HasPrefix(strings.ToUpper(c), ".CLOSE") {
								if CurrentDB != nil {
									CloseDB(CurrentDB)
								} else {
									ui.SetStatus("No database open")
								}
							} else {
								ui.SetStatus(fmt.Sprintf("Unknow command %s", c))
							}
						}
					}
				}
			}
		}
	} else {
		ACmd = append(ACmd, c)
		ui.SetStatus(fmt.Sprintf("Executing %s", c))
		if strings.HasPrefix(strings.ToUpper(c), "SELECT") {
			doSelect(c)
		} else {
			if CurrentDB != nil {
				DoExec(c)
			} else {
				ui.SetStatus("No open database")
			}
		}
	}
	ui.TxtPrompt.SetText("", false)
}

// ****************************************************************************
// DoExec()
// ****************************************************************************
func DoExec(cmd string) {
	_, err := CurrentDB.Exec(cmd)
	if err != nil {
		ui.SetStatus(err.Error())
	} else {
		ui.SetStatus(fmt.Sprintf("Executing %s", cmd))
	}
}

// ****************************************************************************
// OpenDB()
// ****************************************************************************
func OpenDB(fName string) error {
	db, err := sql.Open("sqlite3", fName)
	if err == nil {
		CurrentDB = db
		ui.TxtSQLName.SetText(fmt.Sprintf("Database [yellow]%s", fName))
		CurrentDatabaseName = fName
		showTreeDB()
		ui.SetStatus(fmt.Sprintf("Database %s open successfully", fName))
	}
	return err
}

// ****************************************************************************
// CloseDB()
// ****************************************************************************
func CloseDB(db *sql.DB) {
	db.Close()
	ui.SetStatus("Database closed")
	ui.TxtSQLName.SetText("")
	ui.TblSQLOutput.Clear()
	ui.TblSQLTables.Clear()
	ui.TrvSQLDatabase.GetRoot().ClearChildren()
	root = tview.NewTreeNode("")
	ui.TrvSQLDatabase.SetRoot(root).SetCurrentNode(root)
	CurrentDB = nil
	CurrentDatabaseName = ""
}

// ****************************************************************************
// DoCloseDB()
// ****************************************************************************
func DoCloseDB() {
	DlgCloseDB = DlgCloseDB.YesNoCancel(fmt.Sprintf("Close Database %s", CurrentDatabaseName), // Title
		"This file has been modified. Do you want to save it ?", // Message
		confirmCloseDB,
		0,
		ui.GetCurrentScreen(), ui.TxtPrompt) // Focus return
	ui.PgsApp.AddPage("dlgCloseDB", DlgCloseDB.Popup(), true, false)
	ui.PgsApp.ShowPage("dlgCloseDB")
}

// ****************************************************************************
// confirmCloseDB()
// ****************************************************************************
func confirmCloseDB(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_YES {
		CloseDB(CurrentDB)
	}
}

// ****************************************************************************
// DoOpenDB()
// ****************************************************************************
func DoOpenDB(path string) {
	DlgOpenDB = DlgOpenDB.Input("Open Database", // Title
		"Please, enter the name for the database to open :", // Message
		path,
		confirmOpenDB,
		0,
		ui.GetCurrentScreen(), ui.TxtPrompt) // Focus return
	ui.PgsApp.AddPage("dlgOpenDB", DlgOpenDB.Popup(), true, false)
	ui.PgsApp.ShowPage("dlgOpenDB")
}

// ****************************************************************************
// confirmOpenDB()
// ****************************************************************************
func confirmOpenDB(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_OK {
		newDB := DlgOpenDB.Value
		err := OpenDB(newDB)
		if err != nil {
			ui.SetStatus(err.Error())
		} else {
			ui.SetStatus(fmt.Sprintf("Database %s successfully open", newDB))
		}
	}
}

// ****************************************************************************
// showTreeDB()
// ****************************************************************************
func showTreeDB() {
	root = tview.NewTreeNode(filepath.Base(CurrentDatabaseName))
	ui.TrvSQLDatabase.SetRoot(root).SetCurrentNode(root)
	nodeTables := tview.NewTreeNode("Tables")
	tables := getTables()
	for _, t := range tables {
		nodeTables.AddChild(tview.NewTreeNode(t))
	}
	root.AddChild(nodeTables)
	nodeViews := tview.NewTreeNode("Views")
	views := getViews()
	for _, t := range views {
		nodeViews.AddChild(tview.NewTreeNode(t))
	}
	root.AddChild(nodeViews)
	nodeIndexes := tview.NewTreeNode("Indexes")
	indexes := getIndexes()
	for _, t := range indexes {
		nodeIndexes.AddChild(tview.NewTreeNode(t))
	}
	root.AddChild(nodeIndexes)
	nodeTriggers := tview.NewTreeNode("Triggers")
	triggers := getTriggers()
	for _, t := range triggers {
		nodeTriggers.AddChild(tview.NewTreeNode(t))
	}
	root.AddChild(nodeTriggers)
}

// ****************************************************************************
// getTables()
// ****************************************************************************
func getTables() []string {
	var tables []string
	rows, err := CurrentDB.Query("SELECT name FROM sqlite_schema WHERE type ='table' AND name NOT LIKE 'sqlite_%';")
	if err == nil {
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err == nil {
				tables = append(tables, name)
			}
		}
	}
	return tables
}

// ****************************************************************************
// getViews()
// ****************************************************************************
func getViews() []string {
	var views []string
	rows, err := CurrentDB.Query("SELECT name FROM sqlite_schema WHERE type = 'view';")
	if err == nil {
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err == nil {
				views = append(views, name)
			}
		}
	}
	return views
}

// ****************************************************************************
// getIndexes()
// ****************************************************************************
func getIndexes() []string {
	var indexes []string
	rows, err := CurrentDB.Query("SELECT name FROM sqlite_master WHERE type = 'index';")
	if err == nil {
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err == nil {
				indexes = append(indexes, name)
			}
		}
	}
	return indexes
}

// ****************************************************************************
// getTriggers()
// ****************************************************************************
func getTriggers() []string {
	var triggers []string
	rows, err := CurrentDB.Query("select name from sqlite_master where type = 'trigger';")
	if err == nil {
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err == nil {
				triggers = append(triggers, name)
			}
		}
	}
	return triggers
}

// ****************************************************************************
// doSelect()
// ****************************************************************************
func doSelect(q string) {
	if CurrentDB != nil {
		ui.TblSQLOutput.Clear()
		var myMap = make(map[string]interface{})
		rows, err := CurrentDB.Query(q)
		if err != nil {
			ui.SetStatus(err.Error())
		} else {
			defer rows.Close()
			colNames, err := rows.Columns()
			if err != nil {
				ui.SetStatus(err.Error())
			} else {
				cols := make([]interface{}, len(colNames))
				colPtrs := make([]interface{}, len(colNames))
				for i := 0; i < len(colNames); i++ {
					colPtrs[i] = &cols[i]
				}
				// Header of fields names
				for k, colName := range colNames {
					ui.TblSQLOutput.SetCell(0, k, tview.NewTableCell(tview.Escape("["+colName+"]")).SetAlign(tview.AlignLeft).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
				}
				i := 1
				for rows.Next() {
					err = rows.Scan(colPtrs...)
					if err != nil {
						ui.SetStatus(err.Error())
					} else {
						for k, col := range cols {
							myMap[colNames[k]] = col
						}
						j := 0
						for k := range cols {
							field := colNames[k]
							value := myMap[field]
							if reflect.TypeOf(value) != nil {
								typeVal := reflect.TypeOf(value).String()
								if typeVal == "string" {
									ui.TblSQLOutput.SetCell(i, j, tview.NewTableCell(fmt.Sprintf("%s", value)))
								}
								if strings.HasPrefix(typeVal, "int") {
									ui.TblSQLOutput.SetCell(i, j, tview.NewTableCell(fmt.Sprintf("%d", value)).SetAlign(tview.AlignRight))
								}
								if strings.HasPrefix(typeVal, "float") {
									ui.TblSQLOutput.SetCell(i, j, tview.NewTableCell(fmt.Sprintf("%f", value)).SetAlign(tview.AlignRight))
								}
								if strings.HasPrefix(typeVal, "bool") {
									ui.TblSQLOutput.SetCell(i, j, tview.NewTableCell(fmt.Sprintf("%t", value)).SetAlign(tview.AlignCenter))
								}
							} else {
								ui.TblSQLOutput.SetCell(i, j, tview.NewTableCell("(NULL)").SetAlign(tview.AlignCenter))
							}
							j++
						}
					}
					i++
				}
				ui.TblSQLOutput.SetFixed(1, 0)
				ui.TblSQLOutput.Select(1, 0)
				ui.App.SetFocus(ui.TblSQLOutput)
			}
		}
	} else {
		ui.SetStatus("No open database")
	}
}

// ****************************************************************************
// SetSQLMenu()
// ****************************************************************************
func SetSQLMenu() {
	MnuSQL = MnuSQL.New("Actions", "sqlite3", ui.TblSQLOutput)
	MnuSQL.AddItem("mnuExportCell", "Export cell", DoExportCell, nil, true, false)
	MnuSQL.AddItem("mnuExportRow", "Export row to CSV", DoExportRow, nil, true, false)
	MnuSQL.AddItem("mnuExportAll", "Export all to CSV", DoExportAll, nil, true, false)
	MnuSQL.AddItem("mnuExportStructSQL", "Export structure to SQL script", DoExportAll, nil, true, false)
	MnuSQL.AddItem("mnuExportAllSQL", "Export all to SQL script", DoExportAll, nil, true, false)
	ui.PgsApp.AddPage("dlgMenuAction", MnuSQL.Popup(), true, false)
}

// ****************************************************************************
// ShowMenu()
// ****************************************************************************
func ShowMenu() {
	ui.PgsApp.ShowPage("dlgMenuAction")
}

// ****************************************************************************
// DoExportRow(p any)
// ****************************************************************************
func DoExportRow(p any) {
	r, _ := ui.TblSQLOutput.GetSelection()
	if r > 0 {
		f, err := os.CreateTemp(conf.Cwd, conf.NEW_FILE_TEMPLATE)
		if err == nil {
			defer f.Close()
			w := bufio.NewWriter(f)
			out := ""
			for c := 0; c < ui.TblSQLOutput.GetColumnCount(); c++ {
				out = out + fmt.Sprintf("\"%s\",", ui.TblSQLOutput.GetCell(r, c).Text)
			}
			_, err = fmt.Fprintf(w, "%s", out[:len(out)-1])
			if err != nil {
				ui.SetStatus(err.Error())
			} else {
				w.Flush()
				edit.SwitchToEditor(f.Name())
			}
		} else {
			ui.SetStatus(err.Error())
		}
	}
}

// ****************************************************************************
// DoExportAll(p any)
// ****************************************************************************
func DoExportAll(p any) {
	f, err := os.CreateTemp(conf.Cwd, conf.NEW_FILE_TEMPLATE)
	if err == nil {
		defer f.Close()
		w := bufio.NewWriter(f)
		for r := 0; r < ui.TblSQLOutput.GetRowCount(); r++ {
			out := ""
			for c := 0; c < ui.TblSQLOutput.GetColumnCount(); c++ {
				fldName := ui.TblSQLOutput.GetCell(r, c).Text
				if r == 0 {
					// Special case for escaping '[]' in field's name in first row
					fldName = fldName[:len(fldName)-2] + fldName[len(fldName)-1:]
				}
				out = out + fmt.Sprintf("\"%s\",", fldName)
			}
			_, err = fmt.Fprintf(w, "%s\n", out[:len(out)-1])
			if err != nil {
				ui.SetStatus(err.Error())
			}
		}
		w.Flush()
		edit.SwitchToEditor(f.Name())
	} else {
		ui.SetStatus(err.Error())
	}
}

// ****************************************************************************
// DoExportCell(p any)
// ****************************************************************************
func DoExportCell(p any) {
	r, c := ui.TblSQLOutput.GetSelection()
	if r > 0 {
		f, err := os.CreateTemp(conf.Cwd, conf.NEW_FILE_TEMPLATE)
		if err == nil {
			defer f.Close()
			w := bufio.NewWriter(f)
			_, err = fmt.Fprintf(w, "%s", ui.TblSQLOutput.GetCell(r, c).Text)
			if err != nil {
				ui.SetStatus(err.Error())
			} else {
				w.Flush()
				edit.SwitchToEditor(f.Name())
			}
		} else {
			ui.SetStatus(err.Error())
		}
	}
}

// ****************************************************************************
// SwitchToSQLite3()
// ****************************************************************************
func SwitchToSQLite3() {
	ui.AddNewScreen(ui.ModeSQLite3, nil, nil)
	ui.App.SetFocus(ui.TxtPrompt)
	ui.SetStatus("Switching to [SQLite3]")
}

// ****************************************************************************
// SelfInit()
// ****************************************************************************
func SelfInit(a any) {
	if ui.CurrentMode == ui.ModeFiles {
		idx, _ := ui.TblFiles.GetSelection()
		fName := filepath.Join(conf.Cwd, strings.TrimSpace(ui.TblFiles.GetCell(idx, 2).Text))
		xtype, _ := mimetype.DetectFile(fName)
		if strings.HasSuffix(xtype.String(), "sqlite3") {
			// Is there an open database ?
			if CurrentDB == nil {
				// no, then open the targeted database
				err := OpenDB(fName)
				if err == nil {
					ui.AddNewScreen(ui.ModeSQLite3, nil, nil)
					ui.App.SetFocus(ui.TxtPrompt)
					ui.SetStatus(fmt.Sprintf("Switching to [SQLite3]"))
				} else {
					ui.CurrentMode = ui.ModeSQLite3
					ui.SetTitle("SQLite3")
					ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
					ui.PgsApp.SwitchToPage(ui.GetCurrentScreen())
					ui.App.SetFocus(ui.TxtPrompt)
					ui.SetStatus(err.Error())
				}
			} else {
				// attach the targeted database to the current database
				DoExec(fmt.Sprintf("attach database '%s' as %s", fName, utils.FilenameWithoutExtension(filepath.Base(fName))))
				ui.CurrentMode = ui.ModeSQLite3
				ui.SetTitle("SQLite3")
				ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
				ui.PgsApp.SwitchToPage(ui.GetCurrentScreen())
				ui.App.SetFocus(ui.TxtPrompt)
			}
		} else {
			ui.CurrentMode = ui.ModeSQLite3
			ui.SetTitle("SQLite3")
			ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
			ui.PgsApp.SwitchToPage(ui.GetCurrentScreen())
			ui.App.SetFocus(ui.TxtPrompt)
		}
	} else {
		ui.CurrentMode = ui.ModeSQLite3
		ui.SetTitle("SQLite3")
		ui.LblKeys.SetText(conf.FKEY_LABELS + "\nCtrl+O=Open Ctrl+S=Save")
		ui.PgsApp.SwitchToPage(ui.GetCurrentScreen())
		ui.App.SetFocus(ui.TxtPrompt)
	}
}
