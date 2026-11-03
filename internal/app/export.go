package app

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	configpkg "samouly.fun/bubbleMonitor/internal/config"
	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

func SaveSnapshotCmd(cpu, memory, disk float64, processCount int) tea.Cmd {
	return func() tea.Msg {
		snapshot := struct {
			Timestamp time.Time
			Cpu       float64
			Memory    float64
			Disk      float64
			Processes int
		}{
			Timestamp: time.Now(),
			Cpu:       cpu,
			Memory:    memory,
			Disk:      disk,
			Processes: processCount,
		}

		path, err := configpkg.ResolvePath("", "bubble_snapshot.json")
		if err != nil {
			return msg.ToastMsg{Message: fmt.Sprintf("Path Error: %v", err), Level: data.ToastError, Duration: 3 * time.Second}
		}

		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return msg.ToastMsg{Message: fmt.Sprintf("Failed to open file %s: %v", path, err), Level: data.ToastError, Duration: 3 * time.Second}
		}
		defer func() { _ = file.Close() }()

		encoder := json.NewEncoder(file)
		if encErr := encoder.Encode(snapshot); encErr != nil {
			return msg.ToastMsg{Message: fmt.Sprintf("Failed to write snapshot: %v", encErr), Level: data.ToastError, Duration: 3 * time.Second}
		}

		csvPath, err := configpkg.ResolvePath("", "bubble_snapshot.csv")
		if err != nil {
			csvPath = "bubble_snapshot.csv"
		}
		csvFile, err := os.OpenFile(csvPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer func() { _ = csvFile.Close() }()
			stat, err := csvFile.Stat()
			if err != nil {
				return msg.ToastMsg{Message: fmt.Sprintf("Failed to stat CSV: %v", err), Level: data.ToastError, Duration: 3 * time.Second}
			}
			if stat != nil && stat.Size() == 0 {
				if _, err := csvFile.WriteString("Timestamp,CPU,Memory,Disk,Processes\n"); err != nil {
					return msg.ToastMsg{Message: "Failed to write CSV header: " + err.Error(), Level: data.ToastError, Duration: 3 * time.Second}
				}
			}
			line := fmt.Sprintf("%s,%.2f,%.2f,%.2f,%d\n", snapshot.Timestamp.Format(time.RFC3339), snapshot.Cpu, snapshot.Memory, snapshot.Disk, snapshot.Processes)
			if _, err := csvFile.WriteString(line); err != nil {
				return msg.ToastMsg{Message: "Failed to write CSV data: " + err.Error(), Level: data.ToastError, Duration: 3 * time.Second}
			}
		}

		return msg.ToastMsg{Message: "Snapshot saved to " + path + " & .csv", Level: data.ToastSuccess, Duration: 3 * time.Second}
	}
}
