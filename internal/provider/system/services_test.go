package system

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"samouly.fun/bubbleMonitor/internal/msg"
)

// TestServicesCmd_Returns verifies the command produces a ServicesMsg
// regardless of which init system is running.
func TestServicesCmd_Returns(t *testing.T) {
	start := time.Now()
	raw := ServicesCmd()()
	elapsed := time.Since(start)

	if raw == nil {
		t.Fatal("ServicesCmd returned nil")
	}
	m, ok := raw.(msg.ServicesMsg)
	if !ok {
		t.Fatalf("expected msg.ServicesMsg, got %T", raw)
	}
	t.Logf("ServicesCmd returned %d services in %s", len(m), elapsed)
	// Should be fast — <2s on any reasonable host.
	if elapsed > 2*time.Second {
		t.Errorf("ServicesCmd took %s, want < 2s", elapsed)
	}
}

// TestConnectionsCmd_Returns verifies the command produces a ConnectionsMsg
// and that it doesn't include unix sockets (the "inet" filter).
func TestConnectionsCmd_Returns(t *testing.T) {
	start := time.Now()
	raw := ConnectionsCmd()()
	elapsed := time.Since(start)

	if raw == nil {
		t.Fatal("ConnectionsCmd returned nil")
	}
	m, ok := raw.(msg.ConnectionsMsg)
	if !ok {
		t.Fatalf("expected msg.ConnectionsMsg, got %T", raw)
	}
	t.Logf("ConnectionsCmd returned %d connections in %s", len(m), elapsed)
	if elapsed > 2*time.Second {
		t.Errorf("ConnectionsCmd took %s, want < 2s", elapsed)
	}
	// Verify no unix:// or unix socket addresses leaked through.
	for _, c := range m {
		addr := c.LocalAddr + c.RemoteAddr
		if containsUnixSocket(addr) {
			t.Errorf("unix socket leaked: %s", addr)
		}
	}
}

func TestSystemLogsCmd_Returns(t *testing.T) {
	start := time.Now()
	raw := SystemLogsCmd()()
	elapsed := time.Since(start)

	if raw == nil {
		t.Fatal("SystemLogsCmd returned nil")
	}
	m, ok := raw.(msg.SysLogMsg)
	if !ok {
		t.Fatalf("expected msg.SysLogMsg, got %T", raw)
	}
	t.Logf("SystemLogsCmd returned %d log lines in %s", len(m), elapsed)
}

// containsUnixSocket returns true if the address string looks like a
// unix socket (path-only, no port, or @-prefixed abstract namespace).
func containsUnixSocket(addr string) bool {
	if addr == "" {
		return false
	}
	// unix paths look like "/run/foo.sock" with no colon-separated port
	if len(addr) > 0 && addr[0] == '/' && !hasPort(addr) {
		return true
	}
	// abstract namespace: "@/foo"
	if len(addr) > 0 && addr[0] == '@' {
		return true
	}
	return false
}

func hasPort(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return true
		}
	}
	return false
}

// Suppress unused import lint if the test file gets compiled in a config
// where tea isn't used elsewhere.
var _ tea.Msg
