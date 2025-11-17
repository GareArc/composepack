package process

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
)

// Command describes a process invocation.
type Command struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

// Runner centralizes os/exec usage so we can stub it in tests.
type Runner struct{}

// NewRunner constructs a process runner.
func NewRunner() *Runner {
	return &Runner{}
}

// Run executes a command and returns stdout/stderr once it completes.
func (r *Runner) Run(ctx context.Context, cmd Command) ([]byte, []byte, error) {
	if cmd.Name == "" {
		return nil, nil, errors.New("command name is required")
	}

	command := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	if cmd.Dir != "" {
		command.Dir = cmd.Dir
	}
	if len(cmd.Env) > 0 {
		command.Env = append(os.Environ(), cmd.Env...)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// IsNotFound reports whether the error indicates the command binary was not found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, exec.ErrNotFound)
	}

	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return errors.Is(execErr.Err, exec.ErrNotFound)
	}

	return false
}
