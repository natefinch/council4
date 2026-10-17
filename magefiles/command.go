package magefiles

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func runCommand(ctx context.Context, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("running %s: %w", name, err)
	}
	return nil
}
