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

import "github.com/gdamore/tcell/v2"

const (
	STATUS_MESSAGE_DURATION = 3
	APP_NAME                = "Gosh"
	APP_STRING              = "Gosh © JPL 2023"
	APP_VERSION             = "0.1.0"
	APP_URL                 = "https://github.com/jplozf/gosh"
	HISTORY_CMD_FILE        = "commands_history"
	APP_FOLDER              = ".gosh"
	FILE_MAX_PREVIEW        = 1024
	HASH_THRESHOLD_SIZE     = 1_073_741_824.0
	COLOR_FOLDER            = tcell.ColorLightGreen
	COLOR_FILE              = tcell.ColorYellow
	COLOR_EXECUTABLE        = tcell.ColorLightYellow
	COLOR_SELECTED          = tcell.ColorRed
)
