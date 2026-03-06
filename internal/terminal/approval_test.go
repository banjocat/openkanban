package terminal

import (
	"testing"
)

func TestCommandApprover_Approve(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		approved  bool
		errorCode string
	}{
		{
			name:     "normal command approved",
			command:  "echo",
			args:     []string{"hello"},
			approved: true,
		},
		{
			name:      "sudo command rejected",
			command:   "sudo",
			args:      []string{"apt-get", "install"},
			approved:  false,
			errorCode: "sudo_not_allowed",
		},
		{
			name:      "command with sudo in full path rejected",
			command:   "/usr/bin/sudo",
			args:      []string{"ls"},
			approved:  false,
			errorCode: "sudo_not_allowed",
		},
		{
			name:      "command with sudo in arguments rejected",
			command:   "sh",
			args:      []string{"-c", "sudo ls -la"},
			approved:  false,
			errorCode: "sudo_not_allowed",
		},
		{
			name:     "git command approved",
			command:  "git",
			args:     []string{"clone", "https://example.com/repo.git"},
			approved: true,
		},
		{
			name:     "npm command approved",
			command:  "npm",
			args:     []string{"install"},
			approved: true,
		},
		{
			name:     "python command approved",
			command:  "python",
			args:     []string{"script.py"},
			approved: true,
		},
	}

	ca := NewCommandApprover(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ca.Approve(tt.command, tt.args)

			if result.Approved != tt.approved {
				t.Errorf("got Approved=%v, want %v", result.Approved, tt.approved)
			}

			if !tt.approved && result.Error == nil {
				t.Errorf("expected error, got nil")
			}

			if !tt.approved && tt.errorCode != "" && result.Error.Code != tt.errorCode {
				t.Errorf("got errorCode=%s, want %s", result.Error.Code, tt.errorCode)
			}
		})
	}
}

func TestCommandApprover_ContainsSudo(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains bool
	}{
		{
			name:     "simple arg without sudo",
			args:     []string{"ls", "-la"},
			contains: false,
		},
		{
			name:     "arg with sudo command",
			args:     []string{"sudo apt-get install"},
			contains: true,
		},
		{
			name:     "arg with sudo in middle",
			args:     []string{"-c", "sudo command"},
			contains: true,
		},
		{
			name:     "arg with sudoedit",
			args:     []string{"sudoedit file.txt"},
			contains: false, // sudoedit is not sudo
		},
		{
			name:     "arg with _sudo",
			args:     []string{"my_sudo_function"},
			contains: false,
		},
	}

	ca := NewCommandApprover(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ca.containsSudo(tt.args)
			if result != tt.contains {
				t.Errorf("got containsSudo=%v, want %v", result, tt.contains)
			}
		})
	}
}

func TestCommandApprover_SetAutoApprove(t *testing.T) {
	ca := NewCommandApprover(false)

	if ca.IsAutoApproveEnabled() {
		t.Error("expected IsAutoApproveEnabled to be false initially")
	}

	ca.SetAutoApprove(true)
	if !ca.IsAutoApproveEnabled() {
		t.Error("expected IsAutoApproveEnabled to be true after SetAutoApprove(true)")
	}

	ca.SetAutoApprove(false)
	if ca.IsAutoApproveEnabled() {
		t.Error("expected IsAutoApproveEnabled to be false after SetAutoApprove(false)")
	}
}
