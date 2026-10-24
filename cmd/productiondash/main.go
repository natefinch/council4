// Command productiondash serves a local, auto-refreshing page summarizing every
// council4-production Dynamic Workflow run found in this repository's worktrees.
//
// Runs write their ledgers to .cardwork/workflows/<run-id>/ in whichever
// worktree started them, so the dashboard discovers all worktrees with
// `git worktree list` on every scan.
//
//	go run ./cmd/productiondash            # http://localhost:8765
//	go run ./cmd/productiondash -addr :9000 -repo /path/to/any/council4/worktree
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed index.html
var page []byte

// staleAfter is how long a run without a recorded outcome may go without a
// ledger update before the dashboard stops presenting it as running. Long CI
// waits save no stages, so this is deliberately generous.
const staleAfter = 45 * time.Minute

func main() {
	addr := flag.String("addr", "localhost:8765", "listen address")
	repo := flag.String("repo", ".", "any checkout of the council4 repository")
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	})
	http.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		roots, err := worktrees(*repo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(scan(roots, time.Now()))
	})
	log.Printf("council4 production dashboard on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// worktrees lists the repository's top-level worktrees, excluding the
// per-unit worktrees that production runs create under .cardwork.
func worktrees(repo string) ([]string, error) {
	out, err := exec.Command("git", "-C", repo, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees from %s: %w", repo, err)
	}
	var roots []string
	for _, line := range strings.Split(string(out), "\n") {
		path, ok := strings.CutPrefix(line, "worktree ")
		if ok && !strings.Contains(filepath.ToSlash(path), "/.cardwork/") {
			roots = append(roots, path)
		}
	}
	return roots, nil
}

type task struct {
	TaskType     string `json:"taskType"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	Instructions string `json:"instructions"`
}

type gap struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type card struct {
	OracleID string `json:"oracleId"`
	Name     string `json:"name"`
}

type pullRequest struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type result struct {
	Status   string   `json:"status"`
	Summary  string   `json:"summary"`
	Evidence []string `json:"evidence"`
	Gaps     []gap    `json:"gaps"`
}

type landing struct {
	Commit  string       `json:"commit"`
	CIFixes int          `json:"ciFixes"`
	PR      *pullRequest `json:"pr"`
}

type delivered struct {
	NewlySupported          []card `json:"newlySupported"`
	ExistingSupportVerified []card `json:"existingSupportVerified"`
}

type review struct {
	Verdict  string   `json:"verdict"`
	Findings []string `json:"findings"`
}

type unit struct {
	Index    int      `json:"index"`
	Task     task     `json:"task"`
	Status   string   `json:"status"`
	Reason   string   `json:"reason,omitempty"`
	Result   *result  `json:"result,omitempty"`
	Verdict  *review  `json:"verdict,omitempty"`
	Landing  *landing `json:"landing,omitempty"`
	Prepared *struct {
		Delivered *delivered `json:"delivered,omitempty"`
	} `json:"prepared,omitempty"`
}

type stage struct {
	Label           string `json:"label"`
	TaskType        string `json:"taskType"`
	TaskID          string `json:"taskId"`
	UnitIndex       *int   `json:"unitIndex"`
	Phase           string `json:"phase"`
	Round           *int   `json:"round"`
	RequestedModel  string `json:"requestedModel"`
	ReasoningEffort string `json:"reasoningEffort"`
	StartedAt       string `json:"startedAt"`
	CompletedAt     string `json:"completedAt"`
	Outcome         string   `json:"outcome"`
	Agent           *bool    `json:"agent"`
	Verdict         string   `json:"verdict,omitempty"`
	Findings        []string `json:"findings,omitempty"`
}

type current struct {
	Index int  `json:"index"`
	Task  task `json:"task"`
}

type runError struct {
	Message   string `json:"message"`
	At        string `json:"at"`
	Cancelled bool   `json:"cancelled"`
}

type state struct {
	RunID  string `json:"runId"`
	Policy struct {
		Version string `json:"version"`
	} `json:"policy"`
	Args       map[string]any `json:"args"`
	Stages     []stage        `json:"stages"`
	Units      []unit         `json:"units"`
	Current    *current       `json:"currentUnit"`
	Error      *runError      `json:"error"`
	Outcome    string         `json:"outcome"`
	FinishedAt string         `json:"finishedAt"`
}

type costGroup struct {
	UnitIndex   *int    `json:"unitIndex"`
	TaskType    string  `json:"taskType"`
	TaskID      string  `json:"taskId"`
	Phase       string  `json:"phase"`
	Model       string  `json:"model"`
	ActiveMs    float64 `json:"activeMs"`
	Calls       int     `json:"calls"`
	Agents      int     `json:"agents"`
	ObservedAic float64 `json:"observedAic"`
}

type report struct {
	RuntimeActiveMs    float64     `json:"runtimeActiveMs"`
	RuntimeAic         float64     `json:"runtimeAic"`
	UnallocatedNanoAiu float64     `json:"unallocatedNanoAiu"`
	Reconciliation     string      `json:"reconciliation"`
	Groups             []costGroup `json:"groups"`
	TelemetryError     string      `json:"telemetryError"`
}

// Run is one workflow run as presented by the dashboard.
type Run struct {
	ID             string         `json:"id"`
	Worktree       string         `json:"worktree"`
	Mode           string         `json:"mode"`
	Args           map[string]any `json:"args"`
	PolicyVersion  string         `json:"policyVersion"`
	Status         string         `json:"status"`
	StatusDetail   string         `json:"statusDetail"`
	StartedAt      string         `json:"startedAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	FinishedAt     string         `json:"finishedAt,omitempty"`
	Aic            float64        `json:"aic"`
	ActiveMs       float64        `json:"activeMs"`
	Reconciliation string         `json:"reconciliation"`
	UnallocatedAic float64        `json:"unallocatedAic"`
	TelemetryError string         `json:"telemetryError,omitempty"`
	Units          []unit         `json:"units"`
	CurrentUnit    *current       `json:"currentUnit"`
	Costs          []costGroup    `json:"costs"`
	Stages         []stage        `json:"stages"`
	Error          *runError      `json:"error"`
}

func scan(roots []string, now time.Time) []Run {
	seen := make(map[string]bool)
	runs := []Run{}
	for _, root := range roots {
		dirs, err := os.ReadDir(filepath.Join(root, ".cardwork", "workflows"))
		if err != nil {
			continue
		}
		for _, dir := range dirs {
			if !dir.IsDir() || seen[dir.Name()] {
				continue
			}
			run, err := load(filepath.Join(root, ".cardwork", "workflows", dir.Name()), now)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					log.Printf("skipping %s: %v", dir.Name(), err)
				}
				continue
			}
			run.Worktree = root
			seen[dir.Name()] = true
			runs = append(runs, run)
		}
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt > runs[j].StartedAt })
	return runs
}

func load(dir string, now time.Time) (Run, error) {
	var s state
	updated, err := readJSON(filepath.Join(dir, "state.json"), &s)
	if err != nil {
		return Run{}, err
	}
	var r report
	reported, err := readJSON(filepath.Join(dir, "report.json"), &r)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Run{}, err
	}
	if reported.After(updated) {
		updated = reported
	}
	run := Run{
		ID: s.RunID, Args: s.Args, PolicyVersion: s.Policy.Version, UpdatedAt: updated,
		FinishedAt: s.FinishedAt, Aic: r.RuntimeAic, ActiveMs: r.RuntimeActiveMs,
		Reconciliation: r.Reconciliation, UnallocatedAic: r.UnallocatedNanoAiu / 1e9,
		TelemetryError: r.TelemetryError, Units: s.Units, CurrentUnit: s.Current,
		Costs: r.Groups, Stages: s.Stages, Error: s.Error,
	}
	if run.ID == "" {
		run.ID = filepath.Base(dir)
	}
	run.Mode, _ = s.Args["mode"].(string)
	for _, st := range s.Stages {
		if st.StartedAt != "" && (run.StartedAt == "" || st.StartedAt < run.StartedAt) {
			run.StartedAt = st.StartedAt
		}
	}
	if run.StartedAt == "" {
		run.StartedAt = updated.UTC().Format(time.RFC3339Nano)
	}
	run.Status, run.StatusDetail = status(s, now.Sub(updated))
	return run, nil
}

// status reports the workflow's recorded outcome. Runs written before outcomes
// were recorded, or whose process died, have none; those are only presented as
// running while their ledger is still being updated.
func status(s state, idle time.Duration) (string, string) {
	paused := "Paused or cancelled; resume the same run ID to continue."
	switch {
	case s.Outcome == "scope_finished":
		return "finished", "All requested units settled."
	case s.Outcome == "partial":
		return "partial", "Some units did not land."
	case s.Outcome == "paused_or_cancelled", s.Error != nil && s.Error.Cancelled:
		return "paused", paused
	case s.Error != nil:
		return "error", s.Error.Message
	case idle < staleAfter:
		// Relative times are rendered by the page; keeping this text stable lets
		// the page skip re-rendering runs whose ledgers did not change.
		return "running", "The workflow is still updating its ledger."
	default:
		return "unknown", "No final status recorded (older run, or its workflow process stopped). Check the Workflows view."
	}
}

func readJSON(path string, v any) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", path, err)
	}
	return info.ModTime(), nil
}
