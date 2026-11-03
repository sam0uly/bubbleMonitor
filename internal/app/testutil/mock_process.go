package testutil

import (
	"github.com/shirou/gopsutil/v3/process"

	"samouly.fun/bubbleMonitor/internal/data"
)

type MockProcessProvider struct {
	ProcessesFunc    func(sortBy, direction string) ([]data.ProcessInfo, error)
	ProcessCountFunc func() (int, error)
	OpenFilesFunc    func(pid int32) ([]process.OpenFilesStat, error)
	CmdlineFunc      func(pid int32) (string, error)
	UsernameFunc     func(pid int32) (string, error)
	SuspendFunc      func(pid int32) error
	ResumeFunc       func(pid int32) error
	KillFunc         func(pid int32) error
	SetPriorityFunc  func(pid int32, priority int32) error
}

func (m *MockProcessProvider) Processes(sortBy, direction string) ([]data.ProcessInfo, error) {
	if m.ProcessesFunc != nil {
		return m.ProcessesFunc(sortBy, direction)
	}
	return MockProcesses(), nil
}

func (m *MockProcessProvider) ProcessCount() (int, error) {
	if m.ProcessCountFunc != nil {
		return m.ProcessCountFunc()
	}
	return 150, nil
}

func (m *MockProcessProvider) OpenFiles(pid int32) ([]process.OpenFilesStat, error) {
	if m.OpenFilesFunc != nil {
		return m.OpenFilesFunc(pid)
	}
	return []process.OpenFilesStat{
		{Path: "/proc/1/cmdline"},
		{Path: "/proc/1/environ"},
	}, nil
}

func (m *MockProcessProvider) Cmdline(pid int32) (string, error) {
	if m.CmdlineFunc != nil {
		return m.CmdlineFunc(pid)
	}
	return "/usr/bin/testproc --flag", nil
}

func (m *MockProcessProvider) Username(pid int32) (string, error) {
	if m.UsernameFunc != nil {
		return m.UsernameFunc(pid)
	}
	return "root", nil
}

func (m *MockProcessProvider) Suspend(pid int32) error {
	if m.SuspendFunc != nil {
		return m.SuspendFunc(pid)
	}
	return nil
}

func (m *MockProcessProvider) Resume(pid int32) error {
	if m.ResumeFunc != nil {
		return m.ResumeFunc(pid)
	}
	return nil
}

func (m *MockProcessProvider) Kill(pid int32) error {
	if m.KillFunc != nil {
		return m.KillFunc(pid)
	}
	return nil
}

func (m *MockProcessProvider) SetPriority(pid int32, priority int32) error {
	if m.SetPriorityFunc != nil {
		return m.SetPriorityFunc(pid, priority)
	}
	return nil
}
