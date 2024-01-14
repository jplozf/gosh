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
package help

import "gosh/ui"

// ****************************************************************************
// SetHelp()
// ****************************************************************************
func SetHelp() {
	ui.TxtHelp.SetDynamicColors(true).SetText(`[yellow]		 _____ _____ _____ _____
		|   __|     |   __|  |  |
		|  |  |  |  |__   |     |
		|_____|_____|_____|__|__|
        Copyright jpl@ozf.fr 2024

Gosh is a TUI (Text User Interface) for common management functions on a Linux system.
Gosh is written in Go. The main layout interface is inspired by AS400 text console.
		
[red]Fast access to common functions
[yellow]F1   [white]  :  This help
[yellow]F2   [white]  :  Shell
[yellow]F3   [white]  :  Files Manager
[yellow]F4   [white]  :  Process and Services Manager
[yellow]F5   [white]  :  (refresh)
[yellow]F6   [white]  :  Text Editor
[yellow]F7   [white]  :  Network Manager
[yellow]F8   [white]  :  (special functions)
[yellow]F9   [white]  :  SQLite3 Manager
[yellow]F10   [white] :  Users Manager
[yellow]F11   [white] :  Dashboard
[yellow]F12   [white] :  Exit

[red]Shell [white]!shel

[red]Files Manager [white]!file
[yellow]TAB  [white]  : Move between panels
[yellow]Del  [white]  : Delete the file or folder highlighted or the selection
[yellow]Ins  [white]  : Add the current file or folder to the selection
[yellow]Ctrl+A[white] : Select or unselect all the files and folders in the current folder
[yellow]Ctrl+C[white] : Select or unselect all the files and folders in the current folder

[red]Process and Services Manager [white]!proc

[red]Text Editor [white]!edit

[red]Network Manager

[red]SQLite3 Manager
[yellow]TAB  [white]  : Move between panels
[yellow]Del  [white]  : Delete the file or folder highlighted or the selection
[yellow]Ins  [white]  : Add the current file or folder to the selection
[yellow]Ctrl+A[white] : Select or unselect all the files and folders in the current folder
[yellow]Ctrl+C[white] : Select or unselect all the files and folders in the current folder

[red]Process and Services Manager [white]!proc

[red]Text Editor [white]!edit

[red]Network Manager

[red]SQLite3 Manager

`)
}
