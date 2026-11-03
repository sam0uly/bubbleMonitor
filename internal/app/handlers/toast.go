package handlers

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"samouly.fun/bubbleMonitor/internal/data"
	messages "samouly.fun/bubbleMonitor/internal/msg"
)

func AddToastCmd(msg string, level string) tea.Cmd {
	return func() tea.Msg {
		return messages.ToastMsg{Message: msg, Level: level, Duration: 3 * time.Second}
	}
}

func TickToastCmd(id int64, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return messages.ToastTimeoutMsg{ID: id}
	})
}

func HandleToast(m *data.AppState, msg messages.ToastMsg) tea.Cmd {
	id := m.UI.NextToastID
	m.UI.NextToastID++
	t := data.Toast{
		ID:        id,
		Message:   msg.Message,
		Level:     msg.Level,
		StartTime: time.Now(),
		Duration:  msg.Duration,
	}
	m.UI.Toasts = append(m.UI.Toasts, t)
	return TickToastCmd(id, msg.Duration)
}

func HandleToastTimeout(m *data.AppState, msg messages.ToastTimeoutMsg) {
	for i := 0; i < len(m.UI.Toasts); i++ {
		if m.UI.Toasts[i].ID == msg.ID {
			m.UI.Toasts = append(m.UI.Toasts[:i], m.UI.Toasts[i+1:]...)
			break
		}
	}
}
