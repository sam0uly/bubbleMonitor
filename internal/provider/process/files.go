package process

import (
	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/process"

	"samouly.fun/bubbleMonitor/internal/msg"
)

// FetchOpenFilesCmd retrieves open files for a process
func FetchOpenFilesCmd(pid int32) tea.Cmd {
	return func() tea.Msg {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return msg.OpenFilesMsg{Pid: pid, Err: err}
		}
		files, err := proc.OpenFiles()
		if err != nil {
			return msg.OpenFilesMsg{Pid: pid, Err: err}
		}
		return msg.OpenFilesMsg{Pid: pid, Files: files}
	}
}

// FetchProcessCmdlineCmd lazily fetches cmdline for a process
func FetchProcessCmdlineCmd(pid int32) tea.Cmd {
	return func() tea.Msg {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return msg.ProcessCmdlineMsg{Pid: pid, Cmdline: ""}
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			return msg.ProcessCmdlineMsg{Pid: pid, Cmdline: ""}
		}
		return msg.ProcessCmdlineMsg{Pid: pid, Cmdline: cmdline}
	}
}

// FetchProcessUsernameCmd lazily fetches username for a process
func FetchProcessUsernameCmd(pid int32) tea.Cmd {
	return func() tea.Msg {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return msg.ProcessUsernameMsg{Pid: pid, Username: ""}
		}
		username, err := proc.Username()
		if err != nil {
			return msg.ProcessUsernameMsg{Pid: pid, Username: ""}
		}
		return msg.ProcessUsernameMsg{Pid: pid, Username: username}
	}
}
