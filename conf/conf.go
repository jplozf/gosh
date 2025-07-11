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
package conf

import (
	"os"

	"github.com/gdamore/tcell/v2"
)

const (
	STATUS_MESSAGE_DURATION = 3
	APP_NAME                = "Gosh"
	APP_STRING              = "Gosh © jpl@ozf.fr 2024"
	APP_VERSION             = "0.1.0"
	APP_URL                 = "https://github.com/jplozf/gosh"
	FILE_HISTORY_CMD        = "cmd_history"
	FILE_HISTORY_SQL        = "sql_history"
	APP_FOLDER              = ".gosh"
	FILE_MAX_PREVIEW        = 1024
	HASH_THRESHOLD_SIZE     = 1_073_741_824.0
	COLOR_FOLDER            = tcell.ColorLightGreen
	COLOR_FILE              = tcell.ColorYellow
	COLOR_EXECUTABLE        = tcell.ColorLightYellow
	COLOR_SELECTED          = tcell.ColorRed
	ICON_MODIFIED           = "●"
	NEW_FILE_TEMPLATE       = "gosh_edit_"
	LABEL_PARENT_FOLDER     = "<UP>"
	FILE_LOG                = "gosh.log"
	FILE_CONFIG             = "gosh.json"
	FKEY_LABELS             = "F1=Help F2=Prompt F3=Close F4=Stop Cmd F5=Refresh F6=Previous F7=Next F8=Context Menu F10=Main Menu F12=Exit"
)

var Cwd string
var LogFile *os.File
