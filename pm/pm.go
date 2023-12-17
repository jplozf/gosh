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
// pm is the Process Manager module
// ****************************************************************************

import (
	"bufio"
	"fmt"
	"gosh/menu"
	"gosh/ui"
	"gosh/utils"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
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

// ****************************************************************************
// GLOBALS
// ****************************************************************************
var Processes = make(map[string][]ProcessColumns)
var MnuProcess *menu.Menu
var MnuProcessSort *menu.Menu
var sortColumn = SORT_PID
var sortOrder = SORT_ASCENDING
var headerBackgroundColor = tcell.ColorDarkGreen
var headerTextColor = tcell.ColorYellow
var currentUser string

// ****************************************************************************
// SetProcessMenu()
// ****************************************************************************
func SetProcessMenu() {
	MnuProcess = MnuProcess.New("Actions", "process", ui.TblProcess)
	MnuProcess.AddItem("mnuRenice", "Renice", DoRenice, true)
	MnuProcess.AddItem("mnuPause", "Pause / Resume", DoPause, true)
	MnuProcess.AddItem("mnuKill", "Kill", DoKill, true)
	MnuProcess.AddItem("mnuSendSignal", "Send Signal", DoSendSignal, true)
	ui.PgsApp.AddPage("dlgProcessAction", MnuProcess.Popup(), true, false)

	MnuProcessSort = MnuProcessSort.New("Sort by", "process", ui.TblProcess)
	MnuProcessSort.AddItem("mnuSortPIDA", "PID Ascending", DoSortPIDA, false)
	MnuProcessSort.AddItem("mnuSortPIDD", "PID Descending", DoSortPIDD, true)
	MnuProcessSort.AddItem("mnuSortTimeA", "Time Ascending", DoSortTimeA, true)
	MnuProcessSort.AddItem("mnuSortTimeD", "Time Descending", DoSortTimeD, true)
	MnuProcessSort.AddItem("mnuSortPCPUA", "CPU% Ascending", DoSortCPUA, true)
	MnuProcessSort.AddItem("mnuSortPCPUD", "CPU% Descending", DoSortCPUD, true)
	MnuProcessSort.AddItem("mnuSortPMEMA", "MEM% Ascending", DoSortMEMA, true)
	MnuProcessSort.AddItem("mnuSortPMEMD", "MEM% Descending", DoSortMEMD, true)
	ui.PgsApp.AddPage("dlgProcessSort", MnuProcessSort.Popup(), true, false)
}

// ****************************************************************************
// ShowMenu()
// ****************************************************************************
func ShowMenu() {
	ui.PgsApp.ShowPage("dlgProcessAction")
}

// ****************************************************************************
// ShowMenuSort()
// ****************************************************************************
func ShowMenuSort() {
	ui.PgsApp.ShowPage("dlgProcessSort")
}

// ****************************************************************************
// ShowProcesses()
// ****************************************************************************
func ShowProcesses(user string) {
	currentUser = user
	ui.TxtSelection.Clear()
	ui.PgsApp.SwitchToPage("process")
	ui.TxtProcess.SetText(fmt.Sprintf("Overall CPU usage is %.2f%%", utils.CpuUsage))

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
	ui.TblProcess.SetTitle(fmt.Sprintf("[ %s, sorted by %s ]", user, sorted))
	// PID PRI NI S PCPU PMEM VSZ RSS TIME CMD
	for i, process := range Processes[user] {
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
	}
	ShowUsers()
	ui.TblProcess.SetFixed(1, 0)
	ui.TblProcess.Select(1, 0)
}

// ****************************************************************************
// RefreshMe()
// ****************************************************************************
func RefreshMe() {
	ShowProcesses(currentUser)
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
// DoRenice()
// ****************************************************************************
func DoRenice() {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		ui.SetStatus("Renicing process " + strconv.Itoa(targetPID))
	}
}

// ****************************************************************************
// DoPause()
// ****************************************************************************
func DoPause() {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		state := getProcessState(targetPID)
		if state == "S" {
			if sendSignal(targetPID, "-STOP") == 0 {
				ui.SetStatus("Pausing process " + strconv.Itoa(targetPID))
				showProcessDetails(targetPID)
			} else {
				ui.SetStatus("Unable to pause process " + strconv.Itoa(targetPID))
				showProcessDetails(targetPID)
			}
		} else {
			if state == "T" {
				if sendSignal(targetPID, "-CONT") == 0 {
					ui.SetStatus("Resuming process " + strconv.Itoa(targetPID))
					showProcessDetails(targetPID)
				} else {
					ui.SetStatus("Unable to resume process " + strconv.Itoa(targetPID))
					showProcessDetails(targetPID)
				}
			} else {
				ui.SetStatus("Unknown state for process " + strconv.Itoa(targetPID))
			}
		}
	}
}

// ****************************************************************************
// DoKill()
// ****************************************************************************
func DoKill() {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		ui.SetStatus("Killing process " + strconv.Itoa(targetPID))
	}
}

// ****************************************************************************
// DoKill()
// ****************************************************************************
func DoSendSignal() {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		ui.SetStatus("Sending signal to process " + strconv.Itoa(targetPID))
	}
}

// ****************************************************************************
// ProceedProcessAction()
// ****************************************************************************
func ProceedProcessAction() {
	idx, _ := ui.TblProcess.GetSelection()
	if idx > 0 {
		targetPID, _ := strconv.Atoi(ui.TblProcess.GetCell(idx, 0).Text)
		showProcessDetails(targetPID)
		ui.SetStatus(fmt.Sprintf("Details for process %d", targetPID))
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
// DoSortCPUA()
// ****************************************************************************
func DoSortCPUA() {
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
// DoSortCPUD()
// ****************************************************************************
func DoSortCPUD() {
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
// DoSortMEMA()
// ****************************************************************************
func DoSortMEMA() {
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
// DoSortMEMD()
// ****************************************************************************
func DoSortMEMD() {
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
// DoSortPIDA()
// ****************************************************************************
func DoSortPIDA() {
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
// DoSortPIDD()
// ****************************************************************************
func DoSortPIDD() {
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
// DoSortTimeA()
// ****************************************************************************
func DoSortTimeA() {
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
// DoSortTimeD()
// ****************************************************************************
func DoSortTimeD() {
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
