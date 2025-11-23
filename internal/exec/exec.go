package exec

//go:generate mockgen -source=exec.go -destination=exec_mock.go -package=exec

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Executor interface {
	Run(command string, args []string) error
	RunCapture(command string, args ...string) ([]string, error)
}

type cmdExec struct {
	ctx context.Context
}

func NewExecutor(ctx context.Context) Executor {
	return &cmdExec{
		ctx: ctx,
	}
}

func (e *cmdExec) Run(command string, args []string) error {
	cmd := exec.CommandContext(e.ctx, command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s command failed: %w", command, err)
	}
	return nil
}

func (e *cmdExec) RunCapture(command string, args ...string) ([]string, error) {
	cmd := exec.CommandContext(e.ctx, command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout for %s command: %w", command, err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s command: %w", command, err)
	}

	var lines []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("error reading %s output: %w", command, err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%s command failed: %w", command, err)
	}

	return lines, nil
}
