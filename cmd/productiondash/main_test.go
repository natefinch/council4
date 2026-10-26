package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func writeRun(t *testing.T, root, id, stateJSON, reportJSON string) {
	t.Helper()
	dir := filepath.Join(root, ".cardwork", "workflows", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if stateJSON != "" {
		if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(stateJSON), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if reportJSON != "" {
		if err := os.WriteFile(filepath.Join(dir, "report.json"), []byte(reportJSON), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestScanCollatesRunsAcrossWorktrees(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	writeRun(t, first, "old", `{"runId":"old","policy":{"version":"1.2.0"},"args":{"mode":"produce","maxTasks":4},
		"stages":[{"label":"card/b/unit-0/implementation/0","startedAt":"2026-10-23T04:30:00Z","completedAt":"2026-10-23T04:40:00Z"}],
		"units":[{"index":0,"task":{"taskType":"card","id":"b"},"status":"landed",
			"landing":{"commit":"abc","pr":{"number":3317,"url":"https://example/pr/3317"}},
			"prepared":{"delivered":{"newlySupported":[{"oracleId":"x","name":"Card X"}]}}}],
		"outcome":"scope_finished","finishedAt":"2026-10-23T05:00:00Z"}`,
		`{"runtimeAic":113.9,"runtimeActiveMs":60000,"reconciliation":"matched","unallocatedNanoAiu":0,
		"groups":[{"taskType":"card","taskId":"b","phase":"implementation","model":"gpt-6-luna","observedAic":7.0,"activeMs":1000,"agents":1}]}`)
	writeRun(t, second, "new", `{"runId":"new","policy":{"version":"1.9.0"},"args":{"mode":"produce"},
		"stages":[{"label":"x","startedAt":"2026-10-24T00:53:00Z"},
			{"label":"engine/g/unit-0/review/0","unitIndex":0,"phase":"review","round":0,"verdict":"changes_needed","findings":["wrong controller"]}],
		"units":[{"index":0,"task":{"taskType":"engine","id":"g"},"status":"review_rejected",
			"verdict":{"verdict":"changes_needed","findings":["wrong controller"]}}]}`, "")
	writeRun(t, second, "empty", "", "")
	// The same run reachable from two worktrees is reported once.
	writeRun(t, second, "old", `{"runId":"old","stages":[],"units":[]}`, "")

	shared := t.TempDir()
	if err := os.MkdirAll(filepath.Join(shared, "gone"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shared, "gone", "state.json"), []byte(`{"runId":"gone","stages":[{"label":"x","startedAt":"2026-10-20T00:00:00Z"}],"units":[],"outcome":"scope_finished"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(shared, "old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shared, "old", "state.json"), []byte(`{"runId":"old","stages":[],"units":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	all := scan([]string{first, second, filepath.Join(t.TempDir(), "missing")}, shared, time.Now())
	if len(all) != 3 || all[2].ID != "gone" || all[2].Worktree != "(worktree removed)" || all[2].Status != "finished" {
		t.Fatalf("shared ledger of a removed worktree not shown, or shadowed a worktree copy: %+v", all)
	}
	runs := all[:2]
	if len(runs) != 2 {
		t.Fatalf("got %d runs, want 2: %+v", len(runs), runs)
	}
	if runs[0].ID != "new" || runs[1].ID != "old" {
		t.Fatalf("runs not newest first: %s, %s", runs[0].ID, runs[1].ID)
	}
	old := runs[1]
	if old.Worktree != first || old.Status != "finished" || old.Aic != 113.9 || old.PolicyVersion != "1.2.0" || old.Mode != "produce" {
		t.Fatalf("old run not loaded from its own worktree: %+v", old)
	}
	if got := old.Units[0].Prepared.Delivered.NewlySupported[0].Name; got != "Card X" || old.Units[0].Landing.PR.Number != 3317 {
		t.Fatalf("unit delivery details lost: %+v", old.Units[0])
	}
	if len(old.Costs) != 1 || old.Costs[0].Model != "gpt-6-luna" {
		t.Fatalf("costs lost: %+v", old.Costs)
	}
	if runs[0].Status != "running" {
		t.Fatalf("freshly updated run without an outcome should be running, got %s", runs[0].Status)
	}
	review := runs[0].Stages[1]
	if review.Verdict != "changes_needed" || len(review.Findings) != 1 || review.Findings[0] != "wrong controller" {
		t.Fatalf("stage review verdict lost: %+v", review)
	}
	if v := runs[0].Units[0].Verdict; v == nil || v.Verdict != "changes_needed" || len(v.Findings) != 1 {
		t.Fatalf("unit final verdict lost: %+v", runs[0].Units[0])
	}
}

func TestStatusNeverInventsCompletion(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    state
		idle time.Duration
		want string
	}{
		{"finished", state{Outcome: "scope_finished"}, time.Hour, "finished"},
		{"partial", state{Outcome: "partial"}, time.Hour, "partial"},
		{"error", state{Outcome: "error", Error: &runError{Message: "boom"}}, 0, "error"},
		{"error without outcome", state{Error: &runError{Message: "boom"}}, time.Hour, "error"},
		{"paused", state{Outcome: "paused_or_cancelled", Error: &runError{Cancelled: true}}, time.Hour, "paused"},
		{"cancelled without outcome", state{Error: &runError{Cancelled: true}}, time.Hour, "paused"},
		{"live", state{}, time.Minute, "running"},
		{"stale", state{}, staleAfter + time.Minute, "unknown"},
	} {
		if got, _ := status(tc.s, tc.idle); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestCatalogListsRegisteredWorkflowsWithPrompts(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is required to read the workflow catalog")
	}
	out, err := catalog(".")
	if err != nil {
		t.Fatal(err)
	}
	var workflows []struct {
		Name     string
		Examples []struct{ Prompt string }
	}
	if err := json.Unmarshal(out, &workflows); err != nil {
		t.Fatal(err)
	}
	prompts := map[string]string{}
	for _, w := range workflows {
		for _, e := range w.Examples {
			prompts[w.Name] = e.Prompt
			break
		}
	}
	want := map[string]string{
		"council4-production":        "Run the `council4-production` workflow with `{\"mode\":\"produce\",\"maxTasks\":20}`",
		"council4-production-report": "Run the `council4-production-report` workflow with `{\"runId\":\"<run id>\"}`",
		"council4-engine-triage":     "Run the `council4-engine-triage` workflow with `{}`",
	}
	if len(prompts) != len(want) {
		t.Fatalf("workflows = %v, want %v", prompts, want)
	}
	for name, prompt := range want {
		if prompts[name] != prompt {
			t.Errorf("%s first prompt = %q, want %q", name, prompts[name], prompt)
		}
	}
}
