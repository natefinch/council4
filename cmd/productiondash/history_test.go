package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, content string, modified time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}
}

func TestHistoryListsEveryWorkflowRun(t *testing.T) {
	now := time.Date(2026, 10, 26, 21, 0, 0, 0, time.UTC)
	root, shared := t.TempDir(), t.TempDir()
	writeRun(t, root, "prod", `{"runId":"prod","args":{"mode":"produce"},"outcome":"scope_finished","finishedAt":"2026-10-26T10:00:00Z",
		"stages":[{"label":"x","taskType":"card","phase":"Implementation","startedAt":"2026-10-26T09:00:00Z","completedAt":"2026-10-26T09:30:00Z"}]}`,
		`{"runtimeAic":12.5,"groups":[{"model":"luna","calls":3,"observedAic":10},{"model":"terra","calls":1,"observedAic":2.5}]}`)
	usage := `{"kind":"agent","agentId":"a"}
{"kind":"usage","key":"1","agentId":"a","at":"2026-10-26T20:10:00Z","model":"gpt-6-luna","nanoAiu":1500000000}
{"kind":"usage","key":"1","agentId":"a","at":"2026-10-26T20:10:00Z","model":"gpt-6-luna","nanoAiu":1500000000}
{"kind":"usage","key":"2","agentId":"a","at":"2026-10-26T20:11:00Z","model":"gpt-6.1-terra","nanoAiu":500000000}
{"kind":"usage","key":"3","agen`
	// A recorded run still in progress: its receipts exceed the last native total.
	writeFile(t, filepath.Join(root, ".cardwork/runs/live/run.json"),
		`{"workflow":"council4-engine-triage","runId":"live","startedAt":"2026-10-26T20:00:00Z","phase":"Triage","aic":1,"outcome":null}`, now.Add(-time.Minute))
	writeFile(t, filepath.Join(root, ".cardwork/runs/live/usage.jsonl"), usage, now.Add(-time.Minute))
	// A finished recorded run whose worktree is gone survives in the shared mirror.
	writeFile(t, filepath.Join(shared, "gone/run.json"),
		`{"workflow":"council4-production-report","runId":"gone","startedAt":"2026-10-25T08:00:00Z","finishedAt":"2026-10-25T08:00:03Z","outcome":"finished","aic":0}`, now.Add(-24*time.Hour))
	// Older triage runs without records: one finished, one still running.
	old := filepath.Join(root, ".cardwork/triage/old")
	writeFile(t, filepath.Join(old, "queued-cards.jsonl"), "", now.Add(-48*time.Hour))
	writeFile(t, filepath.Join(old, "usage.jsonl"), usage, now.Add(-47*time.Hour))
	writeFile(t, filepath.Join(old, "report.md"), "# report", now.Add(-46*time.Hour))
	current := filepath.Join(root, ".cardwork/triage/current")
	writeFile(t, filepath.Join(current, "queued-cards.jsonl"), "", now.Add(-20*time.Minute))
	writeFile(t, filepath.Join(current, "usage.jsonl"), usage, now.Add(-30*time.Second))
	writeFile(t, filepath.Join(current, "verdicts.json"), "[]", now.Add(-time.Minute))
	// A recorded run also present as an older triage directory is listed once.
	writeFile(t, filepath.Join(root, ".cardwork/triage/live/usage.jsonl"), usage, now)

	runs := scan([]string{root}, "", now)
	got := map[string]Entry{}
	for _, e := range history([]string{root}, shared, runs, now) {
		if _, dup := got[e.ID]; dup {
			t.Fatalf("run %s listed twice", e.ID)
		}
		got[e.ID] = e
	}
	if len(got) != 5 {
		t.Fatalf("entries = %v", got)
	}
	check := func(id, workflow, status string, aic float64, source string) Entry {
		t.Helper()
		e := got[id]
		if e.Workflow != workflow || e.Status != status || e.Aic != aic || e.AicSource != source {
			t.Errorf("%s = %s %s %v %s, want %s %s %v %s", id, e.Workflow, e.Status, e.Aic, e.AicSource, workflow, status, aic, source)
		}
		return e
	}
	prod := check("prod", "council4-production", "finished", 12.5, "runtime")
	if !prod.Production || len(prod.Models) != 2 || prod.Models[0].Model != "luna" {
		t.Errorf("production models = %+v", prod.Models)
	}
	live := check("live", "council4-engine-triage", "running", 2, "observed")
	if live.Phase != "Triage" || len(live.Models) != 2 || live.Models[0].Aic != 1.5 {
		t.Errorf("live = %+v", live)
	}
	gone := check("gone", "council4-production-report", "finished", 0, "runtime")
	if gone.Worktree != "(worktree removed)" || gone.FinishedAt == "" {
		t.Errorf("gone = %+v", gone)
	}
	finished := check("old", "council4-engine-triage", "finished", 2, "observed")
	if finished.Agents != 1 || finished.FinishedAt == "" || finished.StartedAt != now.Add(-48*time.Hour).Format(time.RFC3339Nano) {
		t.Errorf("old = %+v", finished)
	}
	running := check("current", "council4-engine-triage", "running", 2, "observed")
	if running.FinishedAt != "" || running.Phase != "Triage or Verify" {
		t.Errorf("current = %+v", running)
	}
}
