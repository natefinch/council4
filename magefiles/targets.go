// Package magefiles defines repository automation targets.
package magefiles

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Tools installs the dev tools used by mage, such as golangci-lint.
func Tools(ctx context.Context) error {
	for _, tool := range goTools {
		if err := installTool(ctx, tool); err != nil {
			return err
		}
	}
	return nil
}

var goTools = map[string]string{
	"golangci-lint": "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.2",
}

var bin = filepath.FromSlash("./.bin")

func installTool(ctx context.Context, tool string) error {
	version, ok := goTools[tool]
	if !ok {
		return fmt.Errorf("unknown tool %q", tool)
	}
	// GOBIN has to be an absolute path.
	path, err := filepath.Abs(bin)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", bin, err)
	}
	if err := os.MkdirAll(bin, 0o750); err != nil {
		return fmt.Errorf("failed to create %s: %w", bin, err)
	}
	cmd := exec.CommandContext(ctx, "go", "install", version)
	cmd.Env = append(os.Environ(), "GOBIN="+path)
	stderr, stdout := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w\nStderr: %s\nStdout: %s", version, err, stderr.String(), stdout.String())
	}
	return nil
}

func runTool(ctx context.Context, tool string, args ...string) error {
	cmd := exec.CommandContext(ctx, filepath.Join(bin, tool), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run %s: %w", tool, err)
	}
	return nil
}

// Lint runs golangci-lint on the codebase.
func Lint(ctx context.Context) error {
	err := installTool(ctx, "golangci-lint")
	if err != nil {
		return err
	}

	return runTool(ctx, "golangci-lint", "run")
}
