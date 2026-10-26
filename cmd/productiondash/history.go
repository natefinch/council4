package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry is one run of any repository workflow in the dashboard's history table.
type Entry struct {
	Workflow   string         `json:"workflow"`
	ID         string         `json:"id"`
	Worktree   string         `json:"worktree"`
	Status     string         `json:"status"`
	Detail     string         `json:"detail"`
	Phase      string         `json:"phase"`
	StartedAt  string         `json:"startedAt"`
	FinishedAt string         `json:"finishedAt,omitempty"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	Aic        float64        `json:"aic"`
	AicSource  string         `json:"aicSource"`
	ActiveMs   float64        `json:"activeMs"`
	Agents     int            `json:"agents"`
	Models     []modelCost    `json:"models"`
	Args       map[string]any `json:"args"`
	Production bool           `json:"production"`
}

type modelCost struct {
	Model string  `json:"model"`
	Calls int     `json:"calls"`
	Aic   float64 `json:"aic"`
}

// runRecord is what record.mjs writes for workflows without their own ledger.
type runRecord struct {
	Workflow   string         `json:"workflow"`
	RunID      string         `json:"runId"`
	Args       map[string]any `json:"args"`
	StartedAt  string         `json:"startedAt"`
	FinishedAt string         `json:"finishedAt"`
	Phase      string         `json:"phase"`
	Outcome    string         `json:"outcome"`
	Error      string         `json:"error"`
	Aic        *float64       `json:"aic"`
	ActiveMs   float64        `json:"activeMs"`
	Agents     int            `json:"agents"`
	Models     []modelCost    `json:"models"`
}

type receipt struct {
	Kind    string  `json:"kind"`
	Key     string  `json:"key"`
	AgentID string  `json:"agentId"`
	At      string  `json:"at"`
	Model   *string `json:"model"`
	NanoAiu *int64  `json:"nanoAiu"`
}

const triageWorkflow = "council4-engine-triage"

// history lists every run of every repository workflow: production runs (from
// their ledgers), recorded runs of the other workflows, and triage runs from
// before runs were recorded, reconstructed from their usage receipts.
func history(roots []string, sharedRecords string, production []Run, now time.Time) []Entry {
	seen := make(map[string]bool)
	entries := []Entry{}
	add := func(e Entry) {
		if !seen[e.ID] {
			seen[e.ID] = true
			entries = append(entries, e)
		}
	}
	for _, r := range production {
		add(productionEntry(r, now))
	}
	type source struct{ dir, worktree string }
	var sources []source
	for _, root := range roots {
		sources = append(sources, source{filepath.Join(root, ".cardwork", "runs"), root})
	}
	if sharedRecords != "" {
		sources = append(sources, source{sharedRecords, "(worktree removed)"})
	}
	for _, src := range sources {
		for _, dir := range subdirs(src.dir) {
			if seen[filepath.Base(dir)] {
				continue
			}
			e, err := loadRecord(dir, now)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					log.Printf("skipping %s: %v", dir, err)
				}
				continue
			}
			e.Worktree = src.worktree
			add(e)
		}
	}
	for _, root := range roots {
		for _, dir := range subdirs(filepath.Join(root, ".cardwork", "triage")) {
			if seen[filepath.Base(dir)] {
				continue
			}
			if e, ok := legacyTriage(dir, now); ok {
				e.Worktree = root
				add(e)
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].StartedAt > entries[j].StartedAt })
	return entries
}

func productionEntry(r Run, now time.Time) Entry {
	models := map[string]*modelCost{}
	var order []string
	for _, g := range r.Costs {
		m := g.Model
		if m == "" {
			m = "unknown"
		}
		if models[m] == nil {
			models[m] = &modelCost{Model: m}
			order = append(order, m)
		}
		models[m].Calls += g.Calls
		models[m].Aic += g.ObservedAic
	}
	e := Entry{
		Workflow: "council4-production", ID: r.ID, Worktree: r.Worktree, Status: r.Status, Detail: r.StatusDetail,
		Phase: r.Mode, StartedAt: r.StartedAt, FinishedAt: r.FinishedAt, UpdatedAt: r.UpdatedAt,
		Aic: r.Aic, AicSource: "runtime", ActiveMs: r.ActiveMs, Args: r.Args, Production: true,
	}
	for _, s := range r.Stages {
		if s.StartedAt != "" && s.CompletedAt == "" && r.Status == "running" {
			e.Phase = s.Phase
		}
	}
	for _, m := range order {
		e.Models = append(e.Models, *models[m])
	}
	sortModels(e.Models)
	if e.FinishedAt == "" && r.Status != "running" {
		e.FinishedAt = r.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return e
}

func loadRecord(dir string, now time.Time) (Entry, error) {
	var rec runRecord
	updated, err := readJSON(filepath.Join(dir, "run.json"), &rec)
	if err != nil {
		return Entry{}, err
	}
	receipts, receiptsUpdated := readReceiptsFile(filepath.Join(dir, "usage.jsonl"))
	if receiptsUpdated.After(updated) {
		updated = receiptsUpdated
	}
	e := Entry{
		Workflow: rec.Workflow, ID: rec.RunID, Phase: rec.Phase, StartedAt: rec.StartedAt, FinishedAt: rec.FinishedAt,
		UpdatedAt: updated, ActiveMs: rec.ActiveMs, Agents: rec.Agents, Models: rec.Models, Args: rec.Args,
		AicSource: "runtime",
	}
	if e.ID == "" {
		e.ID = filepath.Base(dir)
	}
	if rec.Aic != nil {
		e.Aic = *rec.Aic
	}
	// A running workflow's native total is only refreshed at phase changes;
	// its receipts show spend since then.
	if observed, models := costs(receipts); rec.Outcome == "" && observed > e.Aic {
		e.Aic, e.Models, e.AicSource = observed, models, "observed"
	}
	switch rec.Outcome {
	case "finished":
		e.Status, e.Detail = "finished", "Finished."
	case "cancelled":
		e.Status, e.Detail = "paused", "Paused or cancelled; resume the same run ID to continue."
	case "error":
		e.Status, e.Detail = "error", rec.Error
	default:
		e.Status, e.Detail = liveStatus(now.Sub(updated))
	}
	return e, nil
}

// legacyTriage reconstructs a triage run that predates run records from the
// files it wrote: indexes at the start, receipts as agents ran, and report.md
// at the end. Its AIC is the sum of its observed usage receipts.
func legacyTriage(dir string, now time.Time) (Entry, bool) {
	files, err := os.ReadDir(dir)
	if err != nil || len(files) == 0 {
		return Entry{}, false
	}
	var first, last time.Time
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		if first.IsZero() || info.ModTime().Before(first) {
			first = info.ModTime()
		}
		if info.ModTime().After(last) {
			last = info.ModTime()
		}
	}
	receipts, _ := readReceiptsFile(filepath.Join(dir, "usage.jsonl"))
	agents := map[string]bool{}
	for _, r := range receipts {
		if r.Kind == "agent" {
			agents[r.AgentID] = true
		}
		if at, err := time.Parse(time.RFC3339Nano, r.At); err == nil && r.At != "" && at.Before(first) {
			first = at
		}
	}
	aic, models := costs(receipts)
	e := Entry{
		Workflow: triageWorkflow, ID: filepath.Base(dir), StartedAt: first.UTC().Format(time.RFC3339Nano),
		UpdatedAt: last, Aic: aic, AicSource: "observed", Agents: len(agents), Models: models,
	}
	if info, err := os.Stat(filepath.Join(dir, "report.md")); err == nil {
		e.Status, e.Detail, e.Phase = "finished", "Older triage run, reconstructed from its files and usage receipts.", "Report"
		e.FinishedAt = info.ModTime().UTC().Format(time.RFC3339Nano)
		return e, true
	}
	e.Status, e.Detail = liveStatus(now.Sub(last))
	switch {
	case exists(filepath.Join(dir, "verified.json")):
		e.Phase = "Report"
	case exists(filepath.Join(dir, "verdicts.json")):
		e.Phase = "Triage or Verify"
	default:
		e.Phase = "Index or Triage"
	}
	if e.Status != "running" {
		e.Detail = "Stopped before writing its report (cancelled or its workflow process ended)."
		e.FinishedAt = last.UTC().Format(time.RFC3339Nano)
	}
	return e, true
}

func liveStatus(idle time.Duration) (string, string) {
	if idle < staleAfter {
		return "running", "The workflow is still updating its run record."
	}
	return "unknown", "No final status recorded (its workflow process may have stopped). Check the Workflows view."
}

func costs(receipts []receipt) (float64, []modelCost) {
	seen := map[string]bool{}
	byModel := map[string]*modelCost{}
	var total float64
	for _, r := range receipts {
		if r.Kind != "usage" || seen[r.Key] {
			continue
		}
		seen[r.Key] = true
		m := "unknown"
		if r.Model != nil {
			m = *r.Model
		}
		if byModel[m] == nil {
			byModel[m] = &modelCost{Model: m}
		}
		byModel[m].Calls++
		if r.NanoAiu != nil {
			aic := float64(*r.NanoAiu) / 1e9
			byModel[m].Aic += aic
			total += aic
		}
	}
	var models []modelCost
	for _, m := range byModel {
		models = append(models, *m)
	}
	sortModels(models)
	return total, models
}

func sortModels(models []modelCost) {
	sort.Slice(models, func(i, j int) bool {
		if models[i].Aic != models[j].Aic {
			return models[i].Aic > models[j].Aic
		}
		return models[i].Model < models[j].Model
	})
}

// readReceiptsFile reads usage receipts, skipping a partially written last line.
func readReceiptsFile(path string) ([]receipt, time.Time) {
	f, err := os.Open(path)
	if err != nil {
		return nil, time.Time{}
	}
	defer f.Close()
	var modified time.Time
	if info, err := f.Stat(); err == nil {
		modified = info.ModTime()
	}
	var receipts []receipt
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var r receipt
		if json.Unmarshal(scanner.Bytes(), &r) == nil {
			receipts = append(receipts, r)
		}
	}
	return receipts, modified
}

func subdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
