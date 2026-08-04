// Package magefiles defines repository automation targets.
package magefiles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func runTool(ctx context.Context, extraEnv []string, tool string, args ...string) error {
	cmd := exec.CommandContext(ctx, filepath.Join(bin, tool), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run %s: %w", tool, err)
	}
	return nil
}

// Lint runs golangci-lint on the codebase.
func Lint(ctx context.Context) error {
	if err := ensureTool(ctx, "golangci-lint"); err != nil {
		return err
	}

	// Point golangci-lint at a per-worktree cache directory so concurrent lint
	// runs across git worktrees don't serialize on a shared cache lock. Paired
	// with allow-parallel-runners in .golangci.toml, this lets many worktrees
	// lint simultaneously without the "parallel golangci-lint is running" error.
	env, err := lintCacheEnv(ctx)
	if err != nil {
		return err
	}
	return runTool(ctx, env, "golangci-lint", "run")
}

// lintCacheEnv returns environment overrides that point golangci-lint at a
// cache directory unique to the current git worktree, keyed by the worktree
// root path. An explicit GOLANGCI_LINT_CACHE in the environment is respected
// and left untouched.
func lintCacheEnv(ctx context.Context) ([]string, error) {
	if _, ok := os.LookupEnv("GOLANGCI_LINT_CACHE"); ok {
		return nil, nil
	}
	root, err := worktreeRoot(ctx)
	if err != nil {
		return nil, err
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(root))
	dir := filepath.Join(base, "council4-golangci-lint", hex.EncodeToString(sum[:8]))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create lint cache dir %s: %w", dir, err)
	}
	return []string{"GOLANGCI_LINT_CACHE=" + dir}, nil
}

// worktreeRoot returns the absolute root of the current git worktree, falling
// back to the working directory when git is unavailable.
func worktreeRoot(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return os.Getwd()
	}
	return strings.TrimSpace(string(out)), nil
}

// ensureTool installs a tool only when the binary is missing or not already at
// the pinned version, avoiding a redundant `go install` on every invocation.
func ensureTool(ctx context.Context, tool string) error {
	if toolAtPinnedVersion(ctx, tool) {
		return nil
	}
	return installTool(ctx, tool)
}

// toolAtPinnedVersion reports whether the installed tool binary reports the
// version pinned in goTools. It returns false when the tool is unknown, the
// binary is missing, its version cannot be determined, or it does not match, so
// the caller falls back to a reinstall.
func toolAtPinnedVersion(ctx context.Context, tool string) bool {
	pinned, ok := goTools[tool]
	if !ok {
		return false
	}
	at := strings.LastIndex(pinned, "@")
	if at < 0 {
		return false
	}
	want := strings.TrimPrefix(pinned[at+1:], "v")
	if want == "" {
		return false
	}
	out, err := exec.CommandContext(ctx, filepath.Join(bin, tool), "version").CombinedOutput()
	if err != nil {
		return false
	}
	// The version is surrounded by spaces in the tool's output (e.g. "has
	// version 2.11.2 built"), so a spaced match avoids treating "2.11.20" as
	// "2.11.2".
	return strings.Contains(string(out), " "+want+" ")
}
