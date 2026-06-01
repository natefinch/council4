package magefiles

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/sh"
)

var goTools = map[string]string{
	"golangci-lint": "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.2",
}

// Tools installs the dev tools used by mage, such as golangci-lint.
func Tools() error {
	for _, tool := range goTools {
		if err := installTool(tool); err != nil {
			return err
		}
	}
	return nil
}

func installTool(tool string) error {
	version, ok := goTools[tool]
	if !ok {
		return fmt.Errorf("unknown tool %q", tool)
	}
	path, err := filepath.Abs("./.bin")
	if err != nil {
		return fmt.Errorf("failed to get absolute path for ./bin: %w", err)
	}
	if err := sh.RunWith(map[string]string{"GOBIN": path}, "go", "install", version); err != nil {
		return fmt.Errorf("failed to install %s: %w", version, err)
	}
	return nil
}

func runTool(tool string, args ...string) error {
	_, err := sh.Exec(nil, os.Stdout, os.Stderr, fmt.Sprintf("./.bin/%s", tool), args...)
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", tool, err)
	}
	return nil
}

// Lint runs golangci-lint on the codebase.
func Lint() error {
	err := installTool("golangci-lint")
	if err != nil {
		return err
	}

	return runTool("golangci-lint", "run")
}
