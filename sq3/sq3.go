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
	"database/sql"
	"fmt"
	"gosh/cmd"
	"gosh/ui"

	_ "github.com/mattn/go-sqlite3"
)

// ****************************************************************************
// sq3 is the SQLite3 interface
// ****************************************************************************

var (
	CurrentDB *sql.DB
)

// ****************************************************************************
// Xeq()
// ****************************************************************************
func Xeq(c string) {
	if c[0] == '!' {
		cmd.Xeq(c)
	} else {
		ui.SetStatus(fmt.Sprintf("Executing %s", c))
	}
}

// ****************************************************************************
// OpenDB()
// ****************************************************************************
func OpenDB(fName string) error {
	db, err := sql.Open("sqlite3", fName)
	if err == nil {
		CurrentDB = db
	}
	return err
}

// ****************************************************************************
// outSQLite3()
// ****************************************************************************
func outSQLite3(fName string) string {
	db, err := sql.Open("sqlite3", fName)
	if err != nil {
		return err.Error()
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_schema WHERE type ='table' AND name NOT LIKE 'sqlite_%';")
	if err != nil {
		return err.Error()
	}
	defer rows.Close()

	zTables := ""
	nTables := 0
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return err.Error()
		}
		nTables++
		zTables += fmt.Sprintf("#%05d > %s\n", nTables, name)
	}
	return fmt.Sprintf("Total tables in SQLite3 database : %d\n", nTables) + zTables
}

// TABLES :
// SELECT name FROM sqlite_schema WHERE type ='table' AND name NOT LIKE 'sqlite_%';
//
// VIEWS :
// SELECT name FROM sqlite_schema WHERE type = 'view';
//
// INDEXES :
// SELECT name FROM sqlite_master WHERE type = 'index';
//
// TRIGGERS :
// select name from sqlite_master where type = 'trigger';
//
