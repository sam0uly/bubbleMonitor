package process

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"samouly.fun/bubbleMonitor/internal/msg"
	"github.com/shirou/gopsutil/v3/process"
)

// ReniceProcessCmdSafe changes the priority of a process by a delta
// delta < 0 increases priority, delta > 0 decreases priority
func ReniceProcessCmdSafe(pid int32, delta int) tea.Cmd {
	return func() tea.Msg {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return msg.PriorityChangeMsg{Pid: pid, Err: err}
		}

		currentNice, err := proc.Nice()
		if err != nil {
			return msg.PriorityChangeMsg{Pid: pid, Err: err}
		}

		if runtime.GOOS == "windows" {
			// Windows Priority Classes
			// We map internal "levels" to PowerShell PriorityClass strings
			// Levels: 0=Idle, 1=BelowNormal, 2=Normal, 3=AboveNormal, 4=High, 5=RealTime
			priorities := []string{"Idle", "BelowNormal", "Normal", "AboveNormal", "High", "RealTime"}

			// Map current Nice value to Index
			// Gopsutil returns Base Priority on Windows
			var currIdx int
			switch {
			case currentNice <= 4:
				currIdx = 0
			case currentNice <= 6:
				currIdx = 1
			case currentNice <= 8:
				currIdx = 2
			case currentNice <= 10:
				currIdx = 3
			case currentNice <= 13:
				currIdx = 4
			default:
				currIdx = 5
			}

			newIdx := currIdx
			if delta < 0 {
				newIdx++ // Step Up
			} else {
				newIdx-- // Step Down
			}

			if newIdx < 0 {
				newIdx = 0
			}
			if newIdx >= len(priorities) {
				newIdx = len(priorities) - 1
			}

			newClass := priorities[newIdx]
			psCmd := fmt.Sprintf("Get-Process -Id %d | foreach { $_.PriorityClass = '%s' }", pid, newClass)

			cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
			if err := cmd.Run(); err != nil {
				return msg.PriorityChangeMsg{Pid: pid, Priority: int32(newIdx), Err: err}
			}
			return msg.PriorityChangeMsg{Pid: pid, Priority: int32(newIdx), Err: nil}
		}

		// Unix Logic
		// Check if renice exists
		if _, err := exec.LookPath("renice"); err != nil {
			return msg.PriorityChangeMsg{Pid: pid, Err: fmt.Errorf("renice command not found")}
		}

		newPrio := currentNice + int32(delta)
		// Clamp typical nice values -20 to 19
		if newPrio < -20 {
			newPrio = -20
		}
		if newPrio > 19 {
			newPrio = 19
		}

		cmd := exec.Command("renice", "-n", fmt.Sprint(newPrio), "-p", fmt.Sprint(pid))
		if err := cmd.Run(); err != nil {
			return msg.PriorityChangeMsg{Pid: pid, Priority: newPrio, Err: err}
		}
		return msg.PriorityChangeMsg{Pid: pid, Priority: newPrio, Err: nil}
	}
}

// SuspendProcessCmd suspends a process
func SuspendProcessCmd(pid int32) tea.Cmd {
	return func() tea.Msg {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return msg.ProcessControlMsg{Pid: pid, Action: "suspend", Err: err}
		}
		err = proc.Suspend()
		return msg.ProcessControlMsg{Pid: pid, Action: "suspend", Err: err}
	}
}

// ResumeProcessCmd resumes a process
func ResumeProcessCmd(pid int32) tea.Cmd {
	return func() tea.Msg {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return msg.ProcessControlMsg{Pid: pid, Action: "resume", Err: err}
		}
		err = proc.Resume()
		return msg.ProcessControlMsg{Pid: pid, Action: "resume", Err: err}
	}
}
