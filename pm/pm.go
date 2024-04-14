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
package pm

// ****************************************************************************
// pm is the Process & Services Manager module
// ****************************************************************************

import (
	"bufio"
	"fmt"
	"gosh/dialog"
	"gosh/menu"
	"gosh/ui"
	"gosh/utils"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/sys/unix"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type Signal struct {
	number int
	name   string
}

type ProcessColumns struct {
	// ps -eo user,s,pid,ppid,pri,ni,pcpu,lstart,time,times,rss,pmem,vsz,cmd --sort +user,-pcpu --no-heading --date-format '%Y%m%d-%H%M%S'
	user     string
	pid      int
	ppid     int
	priority int
	niceness int
	state    string
	started  string
	time     string
	times    int
	rss      int
	vsz      int
	pmem     float64
	pcpu     float64
	command  string
}

type ServiceColumns struct {
	unit        string
	load        string
	active      string
	sub         string
	description string
}

type SortColumn int

const (
	SORT_PID SortColumn = iota
	SORT_TIME
	SORT_PCPU
	SORT_PMEM
)

const (
	SORT_ASCENDING  = 0
	SORT_DESCENDING = 1
)

type ViewType int

const (
	VIEW_PROCESS = iota
	VIEW_SERVICES
)

// ****************************************************************************
// GLOBALS
// ****************************************************************************
var Processes = make(map[string][]ProcessColumns)
var Services []ServiceColumns
var MnuProcess *menu.Menu
var MnuProcessSort *menu.Menu
var MnuService *menu.Menu
var sortColumn = SORT_PID
var sortOrder = SORT_ASCENDING
var headerBackgroundColor = tcell.ColorDarkGreen
var headerTextColor = tcell.ColorYellow
var currentUser string
var DlgRenice *dialog.Dialog
var DlgSendSignal *dialog.Dialog
var DlgKill *dialog.Dialog
var DlgFind *dialog.Dialog
var Signals []Signal
var FindString string
var CurrentView ViewType
var DlgStartService *dialog.Dialog
var DlgStopService *dialog.Dialog
var DlgRestartService *dialog.Dialog
var DlgEnableService *dialog.Dialog
var DlgDisableService *dialog.Dialog

// ****************************************************************************
// SetProcessMenu()
// ****************************************************************************
func SetProcessMenu() {
	MnuProcess = MnuProcess.New("Actions", ui.GetCurrentScreen(), ui.TblProcess)
	MnuProcess.AddItem("mnuRenice", "Renice", DoRenice, nil, true, false)
	MnuProcess.AddItem("mnuPause", "Pause / Resume", DoPause, nil, true, false)
	MnuProcess.AddItem("mnuKill", "Kill", DoKill, nil, true, false)
	MnuProcess.AddItem("mnuSendSignal", "Send Signal", DoSendSignal, nil, true, false)
	ui.PgsApp.AddPage("dlgProcessAction", MnuProcess.Popup(), true, false)

	MnuProcessSort = MnuProcessSort.New("Sort by", ui.GetCurrentScreen(), ui.TblProcess)
	MnuProcessSort.AddItem("mnuSortPIDA", "PID Ascending", DoSortPIDA, nil, false, false)
	MnuProcessSort.AddItem("mnuSortPIDD", "PID Descending", DoSortPIDD, nil, true, false)
	MnuProcessSort.AddItem("mnuSortTimeA", "Time Ascending", DoSortTimeA, nil, true, false)
	MnuProcessSort.AddItem("mnuSortTimeD", "Time Descending", DoSortTimeD, nil, true, false)
	MnuProcessSort.AddItem("mnuSortPCPUA", "CPU% Ascending", DoSortCPUA, nil, true, false)
	MnuProcessSort.AddItem("mnuSortPCPUD", "CPU% Descending", DoSortCPUD, nil, true, false)
	MnuProcessSort.AddItem("mnuSortPMEMA", "MEM% Ascending", DoSortMEMA, nil, true, false)
	MnuProcessSort.AddItem("mnuSortPMEMD", "MEM% Descending", DoSortMEMD, nil, true, false)
	// MnuProcessSort.AddItem("mnuShowServices", "Show Services", DoShowServices, true)
	ui.PgsApp.AddPage("dlgProcessSort", MnuProcessSort.Popup(), true, false)

	MnuService = MnuService.New("Actions", ui.GetCurrentScreen(), ui.TblProcess)
	MnuService.AddItem("mnuStart", "Start", DoStartService, nil, true, false)
	MnuService.AddItem("mnuStop", "Stop", DoStopService, nil, true, false)
	MnuService.AddItem("mnuRestart", "Restart", DoRestartService, nil, true, false)
	MnuService.AddItem("mnuEnable", "Enable", DoEnableService, nil, true, false)
	MnuService.AddItem("mnuDisable", "Disable", DoDisableService, nil, true, false)
	ui.PgsApp.AddPage("dlgServiceAction", MnuService.Popup(), true, false)
}

// ****************************************************************************
// ShowMenu()
// ****************************************************************************
func ShowMenu() {
	if CurrentView == VIEW_PROCESS {
		ui.PgsApp.ShowPage("dlgProcessAction")
	} else {
		ui.PgsApp.ShowPage("dlgServiceAction")
	}
}

// ****************************************************************************
// ShowMenuSort()
// ****************************************************************************
func ShowMenuSort() {
	if CurrentView == VIEW_PROCESS {
		ui.PgsApp.ShowPage("dlgProcessSort")
	} else {
		ui.SetStatus("No sorting available for services")
	}
}

// ****************************************************************************
// ShowProcesses()
// ****************************************************************************
func ShowProcesses(user string) {
	currentUser = user
	ui.TxtSelection.Clear()
	ui.TxtProcess.SetText(fmt.Sprintf("Overall CPU usage is [yellow]%.2f%%[white]", utils.CpuUsage))

	Processes = readProcesses()
	ui.TblProcess.Clear()
	ui.TxtFileInfo.Clear()

	// Column's Header
	ui.TblProcess.SetCell(0, 0, tview.NewTableCell("PID").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 1, tview.NewTableCell("PRI").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 2, tview.NewTableCell("NI").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 3, tview.NewTableCell("S").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 4, tview.NewTableCell("%CPU").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 5, tview.NewTableCell("%MEM").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 6, tview.NewTableCell("VSZ").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 7, tview.NewTableCell("RSS").SetAlign(tview.AlignRight).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 8, tview.NewTableCell("TIME+").SetAlign(tview.AlignCenter).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 9, tview.NewTableCell("CMD").SetAlign(tview.AlignLeft).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))

	var sorted string
	switch sortColumn {
	case SORT_PCPU:
		sorted = "%CPU"
	case SORT_PID:
		sorted = "PID"
	case SORT_PMEM:
		sorted = "%MEM"
	case SORT_TIME:
		sorted = "Time"
	}
	switch sortOrder {
	case SORT_ASCENDING:
		sorted += " Ascending"
	case SORT_DESCENDING:
		sorted += " Descending"
	}
	if FindString != "" {
		ui.TblProcess.SetTitle(fmt.Sprintf("[ %s, filtered on \"%s\", sorted by %s ]", user, FindString, sorted))
	} else {
		ui.TblProcess.SetTitle(fmt.Sprintf("[ %s, sorted by %s ]", user, sorted))
	}
	// PID PRI NI S PCPU PMEM VSZ RSS TIME CMD
	i := 0
	for _, process := range Processes[user] {
		if FindString != "" {
			if strings.Contains(strings.ToUpper(process.command), strings.ToUpper(FindString)) {
				ui.TblProcess.SetCell(i+1, 0, tview.NewTableCell(strconv.Itoa(process.pid)).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorYellow))
				ui.TblProcess.SetCell(i+1, 1, tview.NewTableCell(strconv.Itoa(process.priority)).SetAlign(tview.AlignRight))
				ui.TblProcess.SetCell(i+1, 2, tview.NewTableCell(strconv.Itoa(process.niceness)).SetAlign(tview.AlignRight))
				ui.TblProcess.SetCell(i+1, 3, tview.NewTableCell(process.state))
				ui.TblProcess.SetCell(i+1, 4, tview.NewTableCell(fmt.Sprintf("%.2f%%", process.pcpu)).SetAlign(tview.AlignRight))
				ui.TblProcess.SetCell(i+1, 5, tview.NewTableCell(fmt.Sprintf("%.2f%%", process.pmem)).SetAlign(tview.AlignRight))
				ui.TblProcess.SetCell(i+1, 6, tview.NewTableCell(strconv.Itoa(process.vsz)).SetAlign(tview.AlignRight))
				ui.TblProcess.SetCell(i+1, 7, tview.NewTableCell(strconv.Itoa(process.rss)).SetAlign(tview.AlignRight))
				ui.TblProcess.SetCell(i+1, 8, tview.NewTableCell(process.time))
				ui.TblProcess.SetCell(i+1, 9, tview.NewTableCell(strings.Trim(process.command, " ")).SetAlign(tview.AlignLeft))
				i++
			}
		} else {
			ui.TblProcess.SetCell(i+1, 0, tview.NewTableCell(strconv.Itoa(process.pid)).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorYellow))
			ui.TblProcess.SetCell(i+1, 1, tview.NewTableCell(strconv.Itoa(process.priority)).SetAlign(tview.AlignRight))
			ui.TblProcess.SetCell(i+1, 2, tview.NewTableCell(strconv.Itoa(process.niceness)).SetAlign(tview.AlignRight))
			ui.TblProcess.SetCell(i+1, 3, tview.NewTableCell(process.state))
			ui.TblProcess.SetCell(i+1, 4, tview.NewTableCell(fmt.Sprintf("%.2f%%", process.pcpu)).SetAlign(tview.AlignRight))
			ui.TblProcess.SetCell(i+1, 5, tview.NewTableCell(fmt.Sprintf("%.2f%%", process.pmem)).SetAlign(tview.AlignRight))
			ui.TblProcess.SetCell(i+1, 6, tview.NewTableCell(strconv.Itoa(process.vsz)).SetAlign(tview.AlignRight))
			ui.TblProcess.SetCell(i+1, 7, tview.NewTableCell(strconv.Itoa(process.rss)).SetAlign(tview.AlignRight))
			ui.TblProcess.SetCell(i+1, 8, tview.NewTableCell(process.time))
			ui.TblProcess.SetCell(i+1, 9, tview.NewTableCell(strings.Trim(process.command, " ")).SetAlign(tview.AlignLeft))
			i++
		}
	}
	ShowUsers()
	ui.TblProcess.SetFixed(1, 0)
	ui.TblProcess.Select(1, 0)
	ui.App.SetFocus(ui.TblProcess)
}

// ****************************************************************************
// RefreshMe()
// ****************************************************************************
func RefreshMe() {
	if CurrentView == VIEW_PROCESS {
		ShowProcesses(currentUser)
	} else {
		ShowServices()
	}
	ui.App.SetFocus(ui.TblProcess)
}

// ****************************************************************************
// ShowUsers()
// ****************************************************************************
func ShowUsers() {
	ui.TblProcUsers.Clear()
	type uproc struct {
		user string
		proc int
		pcpu float64
	}
	var ups []uproc
	for user, proc := range Processes {
		var up uproc
		up.user = user
		up.proc = len(proc)
		for _, p := range proc {
			up.pcpu = up.pcpu + p.pcpu
		}
		ups = append(ups, up)
	}
	sort.SliceStable(ups, func(i, j int) bool { return ups[i].proc > ups[j].proc })
	for i, up := range ups {
		if up.user == currentUser {
			ui.TblProcUsers.SetCell(i, 0, tview.NewTableCell(" ▶ "))
		} else {
			ui.TblProcUsers.SetCell(i, 0, tview.NewTableCell("   "))
		}
		ui.TblProcUsers.SetCell(i, 1, tview.NewTableCell(up.user).SetTextColor(tcell.ColorYellow))
		ui.TblProcUsers.SetCell(i, 2, tview.NewTableCell(strconv.Itoa(up.proc)+" process").SetAlign(tview.AlignRight))
		ui.TblProcUsers.SetCell(i, 3, tview.NewTableCell(fmt.Sprintf("%8.2f%%", up.pcpu)).SetAlign(tview.AlignRight))
	}
}

// ****************************************************************************
// readProcesses()
// ****************************************************************************
func readProcesses() map[string][]ProcessColumns {
	var sort = "+user,"
	if sortOrder == SORT_ASCENDING {
		sort += "+"
	} else {
		sort += "-"
	}
	switch sortColumn {
	case SORT_PID:
		sort += "pid"
	case SORT_TIME:
		sort += "times"
	case SORT_PCPU:
		sort += "pcpu"
	case SORT_PMEM:
		sort += "pmem"
	}
	// ps -eo user,s,pid,ppid,pcpu,lstart,time,times,rss,pmem,vsz,cmd --sort +user,-pcpu --no-heading --date-format '%Y%m%d-%H%M%S'
	var processes = make(map[string][]ProcessColumns)
	cmd := exec.Command("ps", "-eo", "pid,user,s,ppid,pri,ni,pcpu,lstart,time,times,rss,pmem,vsz,cmd", "--sort", sort, "--no-heading", "--date-format", "%Y%m%d-%H%M%S")
	bOut, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	out := string(bOut)
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		var p ProcessColumns
		p.pid, err = strconv.Atoi(fields[0])
		if err == nil {
			user := fields[1]
			p.user = user
			p.state = fields[2]
			p.ppid, _ = strconv.Atoi(fields[3])
			p.priority, _ = strconv.Atoi(fields[4])
			p.niceness, _ = strconv.Atoi(fields[5])
			p.pcpu, _ = strconv.ParseFloat(fields[6], 64)
			p.started = fields[7]
			p.time = fields[8]
			p.times, _ = strconv.Atoi(fields[9])
			p.rss, _ = strconv.Atoi(fields[10])
			p.pmem, _ = strconv.ParseFloat(fields[11], 64)
			p.vsz, _ = strconv.Atoi(fields[12])
			p.command = ""
			for i := 13; i < len(fields); i++ {
				p.command = p.command + fields[i] + " "
			}
			p.command = strings.TrimSpace(p.command)

			processes[user] = append(processes[user], p)
		}
	}
	return processes
}

// ****************************************************************************
// DoRenice(p any)
// ****************************************************************************
func DoRenice(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		DlgRenice = DlgRenice.Input("Renice PID", // Title
			fmt.Sprintf("Please, enter the new niceness for process %d :", targetPID), // Message
			"5",
			confirmRenice,
			targetPID,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgRenice", DlgRenice.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgRenice")
	}
	// renice -n 5  -p 8721
}

// ****************************************************************************
// confirmRenice()
// ****************************************************************************
func confirmRenice(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_OK {
		cmd := exec.Command("renice", "-n", DlgRenice.Value, "-p", fmt.Sprintf("%d", idx))
		if err := cmd.Run(); err != nil {
			ui.SetStatus(err.Error())
		} else {
			ui.SetStatus(fmt.Sprintf("PID %d reniced to value %s", idx, DlgRenice.Value))
			showProcessDetails(idx)
		}
	}
}

// ****************************************************************************
// DoPause(p any)
// ****************************************************************************
func DoPause(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		state := getProcessState(targetPID)
		if state == "S" {
			if sendSignal(targetPID, "-SIGSTOP") == 0 {
				ui.SetStatus(fmt.Sprintf("Pausing process %d", targetPID))
				showProcessDetails(targetPID)
			} else {
				ui.SetStatus(fmt.Sprintf("Unable to pause process %d", targetPID))
				showProcessDetails(targetPID)
			}
		} else {
			if state == "T" {
				if sendSignal(targetPID, "-SIGCONT") == 0 {
					ui.SetStatus(fmt.Sprintf("Resuming process %d", targetPID))
					showProcessDetails(targetPID)
				} else {
					ui.SetStatus(fmt.Sprintf("Unable to resume process %d", targetPID))
					showProcessDetails(targetPID)
				}
			} else {
				ui.SetStatus(fmt.Sprintf("Unknown state for process %d", targetPID))
			}
		}
	}
}

// ****************************************************************************
// DoKill(p any)
// ****************************************************************************
func DoKill(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		DlgKill = DlgKill.YesNo("Kill PID", // Title
			fmt.Sprintf("Are you sure you want to kill process %d :", targetPID), // Message
			confirmKill,
			targetPID,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgKill", DlgKill.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgKill")
	}
}

// ****************************************************************************
// confirmKill()
// ****************************************************************************
func confirmKill(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_YES {
		if sendSignal(idx, "-SIGKILL") == 0 {
			ui.SetStatus(fmt.Sprintf("Killing process %d", idx))
			showProcessDetails(idx)
		} else {
			ui.SetStatus(fmt.Sprintf("Unable to kill process %d", idx))
			showProcessDetails(idx)
		}
	}
}

// ****************************************************************************
// DoSendSignal(p any)
// ****************************************************************************
func DoSendSignal(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		var sig []string
		for _, s := range Signals {
			sig = append(sig, fmt.Sprintf("%d) %s", s.number, s.name))
		}
		DlgSendSignal = DlgSendSignal.List("Send signal", // Title
			fmt.Sprintf("Please, select the signal to send to process %d :", targetPID), // Message
			sig[:],
			confirmSendSignal,
			targetPID,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgSendSignal", DlgSendSignal.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgSendSignal")
		ui.SetStatus(fmt.Sprintf("Sending signal to process %d", targetPID))
	}
}

// ****************************************************************************
// confirmSendSignal()
// ****************************************************************************
func confirmSendSignal(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_OK {
		sig := "-" + strings.Split(DlgSendSignal.Value, ") ")[1]
		if sendSignal(idx, sig) == 0 {
			ui.SetStatus(fmt.Sprintf("Signal %s sended to PID %d", sig, idx))
			showProcessDetails(idx)
		} else {
			ui.SetStatus(fmt.Sprintf("Unable to send signal %s to PID %d", sig, idx))
			showProcessDetails(idx)
		}
	}
}

// ****************************************************************************
// ProceedProcessAction()
// ****************************************************************************
func ProceedProcessAction() {
	idx, _ := ui.TblProcess.GetSelection()
	if CurrentView == VIEW_PROCESS {
		if idx > 0 {
			targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
			showProcessDetails(targetPID)
			ui.SetStatus(fmt.Sprintf("Details for process %d", targetPID))
		}
	} else {
		if idx > 0 {
			service := ui.TblProcess.GetCell(idx, 0).Text
			showServiceDetails(service)
			ui.SetStatus(fmt.Sprintf("Details for service %s", service))
		}
	}
}

// ****************************************************************************
// getProcessState()
// ****************************************************************************
func getProcessState(pid int) string {
	// ps -o s= -p <pid>
	var state string
	cmd := exec.Command("ps", "-o", "s=", "-p", fmt.Sprintf("%d", pid))
	bOut, err := cmd.Output()
	if err == nil {
		out := string(bOut)
		scanner := bufio.NewScanner(strings.NewReader(out))
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) != 0 {
				fields := strings.Fields(line)
				state = fields[0]
			}
		}
	} else {
		state = "!"
	}
	return state
}

// ****************************************************************************
// sendSignal()
// ****************************************************************************
func sendSignal(pid int, signal string) int {
	rc := 0
	cmd := exec.Command("kill", signal, fmt.Sprintf("%d", pid))
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			rc = exitError.ExitCode()
		}
	}
	return rc
}

// ****************************************************************************
// getProcessInfo()
// ****************************************************************************
func getProcessInfo(pid int) ProcessColumns {
	var info ProcessColumns
	for _, proc := range Processes {
		for _, p := range proc {
			if p.pid == pid {
				info = p
				break
			}
		}
	}
	return info
}

// ****************************************************************************
// showProcessDetails()
// ****************************************************************************
func showProcessDetails(pid int) {
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "pid,user,s,ppid,pri,ni,pcpu,lstart,time,times,rss,pmem,vsz,cmd", "--no-heading", "--date-format", "%Y%m%d-%H%M%S")
	bOut, err := cmd.Output()
	if err == nil {
		out := string(bOut)
		scanner := bufio.NewScanner(strings.NewReader(out))
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) != 0 {
				fields := strings.Fields(line)
				xmd := ""
				for i := 13; i < len(fields); i++ {
					xmd += fields[i] + " "
				}
				infos := map[string]string{
					"00PID":                 fields[0],
					"01User":                fields[1],
					"02PID Parent":          fields[3],
					"03Priority":            fields[4],
					"04Niceness":            fields[5],
					"05State":               fields[2],
					"06Started":             fields[7],
					"07Time":                fields[8] + " (" + fields[9] + " s)",
					"08CPU":                 fields[6] + " %",
					"09Virtual Memory Size": fields[12] + " kB",
					"10Resident Set Size":   fields[10] + " kB",
					"11Percentage Memory":   fields[11] + " %",
					"12Command":             xmd,
				}
				ui.DisplayMap(ui.TxtProcInfo, infos)
			}
		}
	} else {
		infos := map[string]string{
			"00" + fmt.Sprintf("%d", pid): "This process is no more running.",
		}
		ui.DisplayMap(ui.TxtProcInfo, infos)
	}
}

// ****************************************************************************
// showServiceDetails()
// ****************************************************************************
func showServiceDetails(service string) {
	cmd := exec.Command("systemctl", "status", service)
	bOut, err := cmd.Output()
	if err == nil {
		out := string(bOut)
		scanner := bufio.NewScanner(strings.NewReader(out))
		i := 0
		infos := make(map[string]string)
		for scanner.Scan() {
			line := scanner.Text()
			infos[fmt.Sprintf("%02d", i)] = strings.TrimSpace(line)
			i++
		}
		ui.DisplayMap(ui.TxtProcInfo, infos)
	} else {
		infos := map[string]string{
			"00" + service: "Can't get status about this service.",
		}
		ui.DisplayMap(ui.TxtProcInfo, infos)
	}
}

// ****************************************************************************
// DoSortCPUA(p any)
// ****************************************************************************
func DoSortCPUA(p any) {
	sortColumn = SORT_PCPU
	sortOrder = SORT_ASCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", false)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", true)
	MnuProcessSort.SetEnabled("mnuSortPIDA", true)
	MnuProcessSort.SetEnabled("mnuSortPIDD", true)
	MnuProcessSort.SetEnabled("mnuSortTimeA", true)
	MnuProcessSort.SetEnabled("mnuSortTimeD", true)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoSortCPUD(p any)
// ****************************************************************************
func DoSortCPUD(p any) {
	sortColumn = SORT_PCPU
	sortOrder = SORT_DESCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", true)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", false)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", true)
	MnuProcessSort.SetEnabled("mnuSortPIDA", true)
	MnuProcessSort.SetEnabled("mnuSortPIDD", true)
	MnuProcessSort.SetEnabled("mnuSortTimeA", true)
	MnuProcessSort.SetEnabled("mnuSortTimeD", true)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoSortMEMA(p any)
// ****************************************************************************
func DoSortMEMA(p any) {
	sortColumn = SORT_PMEM
	sortOrder = SORT_ASCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", true)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", false)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", true)
	MnuProcessSort.SetEnabled("mnuSortPIDA", true)
	MnuProcessSort.SetEnabled("mnuSortPIDD", true)
	MnuProcessSort.SetEnabled("mnuSortTimeA", true)
	MnuProcessSort.SetEnabled("mnuSortTimeD", true)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoSortMEMD(p any)
// ****************************************************************************
func DoSortMEMD(p any) {
	sortColumn = SORT_PMEM
	sortOrder = SORT_DESCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", true)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", false)
	MnuProcessSort.SetEnabled("mnuSortPIDA", true)
	MnuProcessSort.SetEnabled("mnuSortPIDD", true)
	MnuProcessSort.SetEnabled("mnuSortTimeA", true)
	MnuProcessSort.SetEnabled("mnuSortTimeD", true)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoSortPIDA(p any)
// ****************************************************************************
func DoSortPIDA(p any) {
	sortColumn = SORT_PID
	sortOrder = SORT_ASCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", true)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", true)
	MnuProcessSort.SetEnabled("mnuSortPIDA", false)
	MnuProcessSort.SetEnabled("mnuSortPIDD", true)
	MnuProcessSort.SetEnabled("mnuSortTimeA", true)
	MnuProcessSort.SetEnabled("mnuSortTimeD", true)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoSortPIDD(p any)
// ****************************************************************************
func DoSortPIDD(p any) {
	sortColumn = SORT_PID
	sortOrder = SORT_DESCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", true)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", true)
	MnuProcessSort.SetEnabled("mnuSortPIDA", true)
	MnuProcessSort.SetEnabled("mnuSortPIDD", false)
	MnuProcessSort.SetEnabled("mnuSortTimeA", true)
	MnuProcessSort.SetEnabled("mnuSortTimeD", true)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoSortTimeA(p any)
// ****************************************************************************
func DoSortTimeA(p any) {
	sortColumn = SORT_TIME
	sortOrder = SORT_ASCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", true)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", true)
	MnuProcessSort.SetEnabled("mnuSortPIDA", true)
	MnuProcessSort.SetEnabled("mnuSortPIDD", true)
	MnuProcessSort.SetEnabled("mnuSortTimeA", false)
	MnuProcessSort.SetEnabled("mnuSortTimeD", true)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoSortTimeD(p any)
// ****************************************************************************
func DoSortTimeD(p any) {
	sortColumn = SORT_TIME
	sortOrder = SORT_DESCENDING
	MnuProcessSort.SetEnabled("mnuSortPCPUA", true)
	MnuProcessSort.SetEnabled("mnuSortPCPUD", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMA", true)
	MnuProcessSort.SetEnabled("mnuSortPMEMD", true)
	MnuProcessSort.SetEnabled("mnuSortPIDA", true)
	MnuProcessSort.SetEnabled("mnuSortPIDD", true)
	MnuProcessSort.SetEnabled("mnuSortTimeA", true)
	MnuProcessSort.SetEnabled("mnuSortTimeD", false)
	ShowProcesses(currentUser)
}

// ****************************************************************************
// DoFindProcess(p any)
// ****************************************************************************
func DoFindProcess(p any) {
	if CurrentView == VIEW_PROCESS {
		DlgFind = DlgFind.Input("Find Process", // Title
			"Please, enter a part of the name to find", // Message
			FindString,
			confirmFind,
			0,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgFind", DlgFind.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgFind")
	} else {
		DlgFind = DlgFind.Input("Find Service", // Title
			"Please, enter a part of the name to find", // Message
			FindString,
			confirmFind,
			0,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgFind", DlgFind.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgFind")
	}
}

// ****************************************************************************
// confirmFind()
// ****************************************************************************
func confirmFind(rc dialog.DlgButton, idx int) {
	if CurrentView == VIEW_PROCESS {
		if rc == dialog.BUTTON_OK {
			FindString = DlgFind.Value
			ShowProcesses(currentUser)
		}
	} else {
		if rc == dialog.BUTTON_OK {
			FindString = DlgFind.Value
			ShowServices()
		}
	}
}

// ****************************************************************************
// SwitchView()
// ****************************************************************************
func SwitchView() {
	if CurrentView == VIEW_PROCESS {
		CurrentView = VIEW_SERVICES
		ui.SetStatus("Switching view to services")
		ShowServices()
	} else {
		CurrentView = VIEW_PROCESS
		ui.SetStatus("Switching view to process")
		ShowProcesses(currentUser)
	}
}

// ****************************************************************************
// ShowServices()
// ****************************************************************************
func ShowServices() {
	ui.TxtSelection.Clear()
	ui.TxtProcess.SetText(fmt.Sprintf("Overall CPU usage is [yellow]%.2f%%[white]", utils.CpuUsage))

	Services = readServices()
	ui.TblProcess.Clear()
	ui.TxtFileInfo.Clear()

	if FindString != "" {
		ui.TblProcess.SetTitle(fmt.Sprintf("[ Services, filtered on \"%s\" ]", FindString))
	} else {
		ui.TblProcess.SetTitle("[ Services ]")
	}

	// Column's Header
	ui.TblProcess.SetCell(0, 0, tview.NewTableCell("UNIT").SetAlign(tview.AlignLeft).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 1, tview.NewTableCell("LOAD").SetAlign(tview.AlignLeft).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 2, tview.NewTableCell("ACTIVE").SetAlign(tview.AlignLeft).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 3, tview.NewTableCell("SUB").SetAlign(tview.AlignLeft).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))
	ui.TblProcess.SetCell(0, 4, tview.NewTableCell("DESCRIPTION").SetAlign(tview.AlignLeft).SetTextColor(headerTextColor).SetBackgroundColor(headerBackgroundColor))

	// UNIT LOAD ACTIVE SUB DESCRIPTION
	i := 0
	for _, service := range Services {
		if FindString != "" {
			if strings.Contains(strings.ToUpper(service.unit), strings.ToUpper(FindString)) {
				ui.TblProcess.SetCell(i+1, 0, tview.NewTableCell(service.unit).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorYellow))
				if service.load == "not-found" {
					ui.TblProcess.SetCell(i+1, 1, tview.NewTableCell(service.load).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorRed))
				} else {
					ui.TblProcess.SetCell(i+1, 1, tview.NewTableCell(service.load).SetAlign(tview.AlignLeft))
				}
				if service.active == "active" {
					ui.TblProcess.SetCell(i+1, 2, tview.NewTableCell(service.active).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorGreen))
				} else {
					ui.TblProcess.SetCell(i+1, 2, tview.NewTableCell(service.active).SetAlign(tview.AlignLeft))
				}
				ui.TblProcess.SetCell(i+1, 3, tview.NewTableCell(service.sub).SetAlign(tview.AlignLeft))
				ui.TblProcess.SetCell(i+1, 4, tview.NewTableCell(service.description).SetAlign(tview.AlignLeft))
				i++
			}
		} else {
			ui.TblProcess.SetCell(i+1, 0, tview.NewTableCell(service.unit).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorYellow))
			if service.load == "not-found" {
				ui.TblProcess.SetCell(i+1, 1, tview.NewTableCell(service.load).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorRed))
			} else {
				ui.TblProcess.SetCell(i+1, 1, tview.NewTableCell(service.load).SetAlign(tview.AlignLeft))
			}
			if service.active == "active" {
				ui.TblProcess.SetCell(i+1, 2, tview.NewTableCell(service.active).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorGreen))
			} else {
				ui.TblProcess.SetCell(i+1, 2, tview.NewTableCell(service.active).SetAlign(tview.AlignLeft))
			}
			ui.TblProcess.SetCell(i+1, 3, tview.NewTableCell(service.sub).SetAlign(tview.AlignLeft))
			ui.TblProcess.SetCell(i+1, 4, tview.NewTableCell(service.description).SetAlign(tview.AlignLeft))
			i++
		}
	}
	ShowUsers()
	ui.TblProcess.SetFixed(1, 0)
	ui.TblProcess.Select(1, 0)
	ui.App.SetFocus(ui.TblProcess)
}

// ****************************************************************************
// readServices()
// ****************************************************************************
func readServices() []ServiceColumns {
	var services []ServiceColumns
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all")
	bOut, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	out := string(bOut)
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 0 {
			if fields[0] == "UNIT" {
				continue
			}
			if fields[0] == "LOAD" {
				break
			}
			var s ServiceColumns
			if fields[0] == "●" {
				s.unit = fields[1]
				s.load = fields[2]
				s.active = fields[3]
				s.sub = fields[4]
				s.description = ""
				for i := 5; i < len(fields); i++ {
					s.description = s.description + fields[i] + " "
				}
				s.description = strings.TrimSpace(s.description)
			} else {
				s.unit = fields[0]
				s.load = fields[1]
				s.active = fields[2]
				s.sub = fields[3]
				s.description = ""
				for i := 4; i < len(fields); i++ {
					s.description = s.description + fields[i] + " "
				}
				s.description = strings.TrimSpace(s.description)
			}
			services = append(services, s)
		}
	}
	return services
}

/*
ps -eo pid,user,pcpu,lstart,cmd
ps -p <pid> -o lstart
kill -STOP <pid> => pause
kill -CONT <pid> => resume


https://linuxconfig.org/ps-output-difference-between-vsz-vs-rss-memory-usage

VSZ is Virtual Memory Size. This is the size of memory that Linux has given to a process, but it doesn’t necessarily mean that the process is using all of that memory. For example, many applications have functions to carry out certain tasks, but may not load them into memory until they are needed. Linux utilizes demand paging, which only loads pages into memory once the application makes an attempt to use them.

The VSZ size you see has taken all of these pages into consideration, but it doesn’t mean they’ve been loaded into physical memory. The VSZ size is therefore not usually an accurate measurement of how much memory a process is using, but rather an indication of the maximum amount of memory a process can use if it loads all of its functions and libraries into physical memory.

RSS is Resident Set Size. This is the size of memory that a process has currently used to load all of its pages. At first glance, it may seem like the RSS number is the real amount of physical memory that a system process is using. However, shared libraries are counted for each process, making the reported amount of physical memory usage less accurate.

Here’s an example. If you have two image editing programs on your Linux system, they likely utilize many of the same image processing libraries. If you open one of the applications, the necessary library will be loaded into RAM. When you open the second application, it will avoid reloading a duplicate copy of the library into RAM, and just share the same copy that the first application is using. For both applications, the RSS column would count the size of the shared library, even though it’s only been loaded once. This means that the RSS size is often an overestimate of the amount of physical memory that’s actually being used by a process.
*/

// ****************************************************************************
// InitSignals()
// ****************************************************************************
func InitSignals() {
	// https://stackoverflow.com/questions/42598522/how-can-i-list-available-operating-system-signals-by-name-in-a-cross-platform-wa
	for i := syscall.Signal(0); i < syscall.Signal(255); i++ {
		name := unix.SignalName(i)
		// Signal numbers are not guaranteed to be contiguous.
		if name != "" {
			Signals = append(Signals, Signal{int(i), name})
		}
	}
}

// ****************************************************************************
// DoStartService(p any)
// ****************************************************************************
func DoStartService(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		target := ui.TblProcess.GetCell(idx, 0).Text
		DlgStartService = DlgStartService.YesNo("Start Service", // Title
			fmt.Sprintf("Are you sure you want to start service %s :", target), // Message
			confirmStartService,
			idx,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgStartService", DlgStartService.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgStartService")
	}
}

// ****************************************************************************
// confirmStartService()
// ****************************************************************************
func confirmStartService(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_YES {
		service := ui.TblProcess.GetCell(idx, 0).Text
		cmd := exec.Command("sudo", "systemctl", "start", service)
		if err := cmd.Run(); err == nil {
			ui.SetStatus(fmt.Sprintf("Service %s started", service))
			showServiceDetails(service)
		} else {
			ui.SetStatus(err.Error())
			showServiceDetails(service)
		}
	}
}

// ****************************************************************************
// DoStopService(p any)
// ****************************************************************************
func DoStopService(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		target := ui.TblProcess.GetCell(idx, 0).Text
		DlgStopService = DlgStopService.YesNo("Stop Service", // Title
			fmt.Sprintf("Are you sure you want to stop service %s :", target), // Message
			confirmStopService,
			idx,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgStopService", DlgStopService.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgStopService")
	}
}

// ****************************************************************************
// confirmStopService()
// ****************************************************************************
func confirmStopService(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_YES {
		service := ui.TblProcess.GetCell(idx, 0).Text
		cmd := exec.Command("sudo", "systemctl", "stop", service)
		if err := cmd.Run(); err == nil {
			ui.SetStatus(fmt.Sprintf("Service %s stopped", service))
			showServiceDetails(service)
		} else {
			ui.SetStatus(err.Error())
			showServiceDetails(service)
		}
	}
}

// ****************************************************************************
// DoRestartService(p any)
// ****************************************************************************
func DoRestartService(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		target := ui.TblProcess.GetCell(idx, 0).Text
		DlgRestartService = DlgRestartService.YesNo("Restart Service", // Title
			fmt.Sprintf("Are you sure you want to restart service %s :", target), // Message
			confirmRestartService,
			idx,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgRestartService", DlgRestartService.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgRestartService")
	}
}

// ****************************************************************************
// confirmRestartService()
// ****************************************************************************
func confirmRestartService(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_YES {
		service := ui.TblProcess.GetCell(idx, 0).Text
		cmd := exec.Command("sudo", "systemctl", "restart", service)
		if err := cmd.Run(); err == nil {
			ui.SetStatus(fmt.Sprintf("Service %s restarted", service))
			showServiceDetails(service)
		} else {
			ui.SetStatus(err.Error())
			showServiceDetails(service)
		}
	}
}

// ****************************************************************************
// DoEnableService(p any)
// ****************************************************************************
func DoEnableService(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		target := ui.TblProcess.GetCell(idx, 0).Text
		DlgEnableService = DlgEnableService.YesNo("Enable Service", // Title
			fmt.Sprintf("Are you sure you want to enable service %s :", target), // Message
			confirmEnableService,
			idx,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgEnableService", DlgEnableService.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgEnableService")
	}
}

// ****************************************************************************
// confirmEnableService()
// ****************************************************************************
func confirmEnableService(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_YES {
		service := ui.TblProcess.GetCell(idx, 0).Text
		cmd := exec.Command("sudo", "systemctl", "enable", service)
		if err := cmd.Run(); err == nil {
			ui.SetStatus(fmt.Sprintf("Service %s enabled", service))
			showServiceDetails(service)
		} else {
			ui.SetStatus(err.Error())
			showServiceDetails(service)
		}
	}
}

// ****************************************************************************
// DoDisableService(p any)
// ****************************************************************************
func DoDisableService(p any) {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		target := ui.TblProcess.GetCell(idx, 0).Text
		DlgDisableService = DlgDisableService.YesNo("Disable Service", // Title
			fmt.Sprintf("Are you sure you want to disable service %s :", target), // Message
			confirmDisableService,
			idx,
			ui.GetCurrentScreen(), ui.TblProcess) // Focus return
		ui.PgsApp.AddPage("dlgDisableService", DlgDisableService.Popup(), true, false)
		ui.PgsApp.ShowPage("dlgDisableService")
	}
}

// ****************************************************************************
// confirmDisableService()
// ****************************************************************************
func confirmDisableService(rc dialog.DlgButton, idx int) {
	if rc == dialog.BUTTON_YES {
		service := ui.TblProcess.GetCell(idx, 0).Text
		cmd := exec.Command("sudo", "systemctl", "disable", service)
		if err := cmd.Run(); err == nil {
			ui.SetStatus(fmt.Sprintf("Service %s disabled", service))
			showServiceDetails(service)
		} else {
			ui.SetStatus(err.Error())
			showServiceDetails(service)
		}
	}
}

// ****************************************************************************
// SelfInit()
// ****************************************************************************
func SelfInit(user any) {
	SetProcessMenu()
	ShowProcesses(user.(string))
	ui.App.Sync()
	ui.App.SetFocus(ui.TblProcess)
}
