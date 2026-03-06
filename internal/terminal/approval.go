package terminal

import (
	"strings"
)

// CommandApprovalError represents an error during command approval
type CommandApprovalError struct {
	Code    string // error code (e.g., "sudo_not_allowed")
	Message string
}

func (e *CommandApprovalError) Error() string {
	return e.Message
}

// ApprovalResult contains the result of command approval
type ApprovalResult struct {
	Approved bool                  // Whether the command is approved
	Error    *CommandApprovalError // Error if not approved
	Modified bool                  // Whether the command was modified
	Command  string                // The command to execute (may be modified)
	Args     []string              // The arguments (may be modified)
}

// CommandApprover validates and potentially modifies commands before execution
type CommandApprover struct {
	autoApprove bool // Whether to auto-approve commands
}

// NewCommandApprover creates a new command approver
func NewCommandApprover(autoApprove bool) *CommandApprover {
	return &CommandApprover{
		autoApprove: autoApprove,
	}
}

// Approve validates a command before execution
// It checks for sudo and other dangerous commands
func (ca *CommandApprover) Approve(command string, args []string) *ApprovalResult {
	result := &ApprovalResult{
		Approved: true,
		Command:  command,
		Args:     args,
	}

	// Check if command is sudo
	if command == "sudo" || strings.HasSuffix(command, "/sudo") {
		result.Approved = false
		result.Error = &CommandApprovalError{
			Code:    "sudo_not_allowed",
			Message: "sudo commands are not allowed for auto-approval",
		}
		return result
	}

	// Check if sudo appears in arguments (e.g., "sh -c sudo command")
	if ca.containsSudo(args) {
		result.Approved = false
		result.Error = &CommandApprovalError{
			Code:    "sudo_not_allowed",
			Message: "sudo commands are not allowed for auto-approval",
		}
		return result
	}

	return result
}

// containsSudo checks if any argument contains sudo
func (ca *CommandApprover) containsSudo(args []string) bool {
	for _, arg := range args {
		// Check for sudo as standalone word or at start of command string
		// This catches "sudo command" in shell arguments
		parts := strings.Fields(arg)
		for _, part := range parts {
			if part == "sudo" || strings.HasPrefix(part, "sudo ") {
				return true
			}
		}
	}
	return false
}

// SetAutoApprove enables or disables auto-approval mode
func (ca *CommandApprover) SetAutoApprove(enabled bool) {
	ca.autoApprove = enabled
}

// IsAutoApproveEnabled returns whether auto-approval is enabled
func (ca *CommandApprover) IsAutoApproveEnabled() bool {
	return ca.autoApprove
}

// IsPermissionPrompt checks if text contains permission/confirmation prompts
func (ca *CommandApprover) IsPermissionPrompt(text string) bool {
	text = strings.ToLower(text)

	// Keywords that indicate a permission prompt
	prompts := []string{
		"permission",
		"approve",
		"allow",
		"accept",
		"proceed",
		"confirm",
		"[y/n]",
		"(y/n)",
		"[yes/no]",
		"(yes/no)",
		"waiting for",
		"continue?",
		"sure?",
	}

	for _, prompt := range prompts {
		if strings.Contains(text, prompt) {
			return true
		}
	}

	return false
}
