package exec

import (
	"context"
	"strings"
	"testing"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		expectErr bool
	}{
		{
			name:    "True",
			command: "true",
		},
		{
			name:      "False",
			command:   "false",
			expectErr: true,
		},
		{
			name:      "CommandNotFound",
			command:   "blah-blah-blah-this-command-does-not-exist",
			expectErr: true,
		},
		{
			name:    "ShellExit0",
			command: "sh",
			args:    []string{"-c", "exit 0"},
		},
		{
			name:      "ShellExit42",
			command:   "sh",
			args:      []string{"-c", "exit 42"},
			expectErr: true,
		},
		{
			name:      "ShellKillSigPipe",
			command:   "sh",
			args:      []string{"-c", "kill -PIPE $$ || :"},
			expectErr: true,
		},
		{
			name:      "InvalidArgument",
			command:   "sh",
			args:      []string{"-c", "exit 0;\000"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ce := NewExecutor(ctx)
			err := ce.Run(tt.command, tt.args)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestRunCommandCapture(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		expectErr bool
		expectOut []string
	}{
		{
			name:      "EchoSingle",
			command:   "echo",
			args:      []string{"\t\r\fHello, World!  "},
			expectOut: []string{"\t\r\fHello, World!  "},
		},
		{
			name:      "EchoMultiple",
			command:   "sh",
			args:      []string{"-c", "echo line1; echo line2; echo line3"},
			expectOut: []string{"line1", "line2", "line3"},
		},
		{
			name:      "CommandNotFound",
			command:   "blah-blah-blah-this-command-does-not-exist",
			expectErr: true,
		},
		{
			name:      "EmptyOutput",
			command:   "true",
			expectOut: []string{},
		},
		{
			name:      "FailingCommand",
			command:   "sh",
			args:      []string{"-c", "echo 'BOOM!' && exit 1"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ce := NewExecutor(ctx)
			output, err := ce.RunCapture(tt.command, tt.args...)

			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tt.expectErr, err)
			}

			if !strings.EqualFold(strings.Join(output, "\n"), strings.Join(tt.expectOut, "\n")) {
				t.Errorf("Expected output: %v, but got: %v", tt.expectOut, output)
			}
		})
	}
}
