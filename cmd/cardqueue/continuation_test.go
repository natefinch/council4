package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/internal/supportinventory"
	"github.com/natefinch/council4/mtg/cards"
)

func continuationFixture(t *testing.T) (manifest, reviewedPlan, []cardgen.ScryfallCard, supportinventory.Inventory) {
	t.Helper()
	q, corpus, inventory := queueFixture(t)
	q.CorpusSHA256 = "pinned-corpus"
	if err := transition(&q, "claim", "first", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	for _, id := range q.Batches[0].OracleIDs {
		inventory.Production = append(inventory.Production, cards.Support{
			OracleID: id, Loadable: true, BehaviorVerified: true, Verification: "executed exact-card scenario",
		})
	}
	inventory.Backlog = slices.DeleteFunc(inventory.Backlog, func(r supportinventory.Record) bool {
		return slices.Contains(q.Batches[0].OracleIDs, r.OracleID)
	})
	if err := transition(&q, "verify", "first", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	b := q.Batches[1]
	b.ID, b.Review, b.PrerequisiteEvidence = "third", "reviewed source/runtime packet", nil
	plan := reviewedPlan{CorpusSHA256: q.CorpusSHA256, Batches: []batch{b}}
	for i := range q.Records {
		r := &q.Records[i]
		if r.Batch == "second" {
			r.Batch = ""
			plan.Records = append(plan.Records, reviewedRecord{OracleID: r.OracleID, Revision: r.Revision, Name: r.Name})
		}
	}
	q.Batches = q.Batches[:1]
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	return q, plan, corpus, inventory
}

func snapshot(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestReviewedContinuationAfterExhaustionAndResume(t *testing.T) {
	q, plan, corpus, inventory := continuationFixture(t)
	history := snapshot(t, q.Batches)
	baseline := snapshot(t, q.BaselineProduction)
	if next, err := nextBatch(q, "writer"); err != nil || next != nil {
		t.Fatal("fixture must exhaust saved batches", err)
	}
	if candidates := unreviewedCandidates(q); len(candidates) != 3 || candidates[0].OracleID != "d" {
		t.Fatalf("missing bounded exact unassigned candidates: %+v", candidates)
	}
	if err := continueQueue(&q, "add-reviewed-batch", "third", "writer", plan); err != nil {
		t.Fatal(err)
	}
	if snapshot(t, q.Batches[:1]) != history || snapshot(t, q.BaselineProduction) != baseline {
		t.Fatal("continuation rewrote original batch history or baseline")
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if next, err := nextBatch(q, "writer"); err != nil || next != nil {
		t.Fatal("method names or issue numbers alone made the batch ready", err)
	}
	if pending := pendingPrerequisites(q); len(pending) != 1 || pending[0].Batch != "third" || !slices.Equal(pending[0].Issues, []int{3274}) {
		t.Fatalf("pending canonical prerequisites not visible: %+v", pending)
	}
	before := snapshot(t, q)
	if err := transition(&q, "claim", "third", "writer", 0, "", inventory); err == nil || snapshot(t, q) != before {
		t.Fatal("pending prerequisite was claimable or failed claim mutated state")
	}
	evidencePlan := plan
	evidencePlan.Records = nil
	evidencePlan.Batches = slices.Clone(plan.Batches)
	evidencePlan.Batches[0].PrerequisiteEvidence = []prerequisiteEvidence{{
		Issue: 3274, RuntimeRevision: "checked-commit", Evidence: "Search Engine path; TestActualSearch passed",
	}}
	if err := continueQueue(&q, "record-prerequisites", "third", "writer", evidencePlan); err != nil {
		t.Fatal(err)
	}
	if len(pendingPrerequisites(q)) != 0 {
		t.Fatal("checked prerequisite remains pending")
	}
	if err := transition(&q, "claim", "third", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	var resumed manifest
	if err := json.Unmarshal([]byte(snapshot(t, q)), &resumed); err != nil {
		t.Fatal(err)
	}
	if err := validate(resumed, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if next, err := nextBatch(resumed, "writer"); err != nil || next == nil || next.ID != "third" {
		t.Fatal("durable owner claim not resumed", err)
	}
	if err := transition(&resumed, "claim", "third", "writer", 0, "", inventory); err != nil {
		t.Fatal("same-owner claim is not idempotent", err)
	}
	for _, owner := range []string{"", "other"} {
		before := snapshot(t, resumed)
		if _, err := nextBatch(resumed, owner); err == nil {
			t.Fatal("another owner selected the saved claim")
		}
		if err := transition(&resumed, "claim", "third", owner, 0, "", inventory); err == nil || snapshot(t, resumed) != before {
			t.Fatal("another owner stole or altered the claim")
		}
	}
	for _, owner := range []string{"writer", "other"} {
		before := snapshot(t, resumed)
		if err := continueQueue(&resumed, "add-reviewed-batch", "third", owner, plan); err == nil || snapshot(t, resumed) != before {
			t.Fatal("replanning silently replaced an active claim")
		}
	}
}

func TestRecordBlockersMarksUnbatchedCardsAndSchedulesTheGap(t *testing.T) {
	q, plan, corpus, inventory := continuationFixture(t)
	blocked := plan.Records[0]
	blocked.Blocker, blocked.Reason = 4100, "triage: needs combined search"
	blockers := reviewedPlan{CorpusSHA256: q.CorpusSHA256, Records: []reviewedRecord{blocked}}
	for name, mutate := range map[string]func(*reviewedPlan) (string, string){
		"no reason":  func(p *reviewedPlan) (string, string) { p.Records[0].Reason = ""; return "", "writer" },
		"no blocker": func(p *reviewedPlan) (string, string) { p.Records[0].Blocker = 0; return "", "writer" },
		"batch flag": func(p *reviewedPlan) (string, string) { return "third", "writer" },
		"no owner":   func(p *reviewedPlan) (string, string) { return "", "" },
		"stale":      func(p *reviewedPlan) (string, string) { p.Records[0].Revision = "old"; return "", "writer" },
	} {
		bad := blockers
		bad.Records = slices.Clone(blockers.Records)
		batch, owner := mutate(&bad)
		before := snapshot(t, q)
		if err := continueQueue(&q, "record-blockers", batch, owner, bad); err == nil || snapshot(t, q) != before {
			t.Fatalf("%s: invalid blocker plan accepted or changed state", name)
		}
	}
	if err := continueQueue(&q, "record-blockers", "", "writer", blockers); err != nil {
		t.Fatal(err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	pending := pendingPrerequisites(q)
	if len(pending) != 1 || !slices.Equal(pending[0].Issues, []int{4100}) || pending[0].OracleIDs[0] != blocked.OracleID {
		t.Fatalf("triage blocker not scheduled: %+v", pending)
	}
	if err := continueQueue(&q, "record-blockers", "", "writer", blockers); err == nil {
		t.Fatal("already blocked card re-blocked")
	}
}

func TestBatchWithoutPrerequisitesIsReadyOnExistingOperations(t *testing.T) {
	q, plan, corpus, inventory := continuationFixture(t)
	plan.Batches[0].Prerequisites = nil
	if err := continueQueue(&q, "add-reviewed-batch", "third", "writer", plan); err != nil {
		t.Fatal(err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if next, err := nextBatch(q, "writer"); err != nil || next == nil || next.ID != "third" {
		t.Fatalf("batch using only existing operations is not claimable: %+v, %v", next, err)
	}
	if pending := pendingPrerequisites(q); len(pending) != 0 {
		t.Fatalf("empty prerequisites reported pending: %+v", pending)
	}
}

func TestInvalidBatchErrorsExplainWhy(t *testing.T) {
	for want, b := range map[string]batch{
		"missing id":     {Family: "f", OracleIDs: []string{"a", "b", "c"}, Operations: []string{"op"}},
		"missing family": {ID: "x", OracleIDs: []string{"a", "b", "c"}, Operations: []string{"op"}},
		"needs 3–5":      {ID: "x", Family: "f", OracleIDs: []string{"a", "b"}, Operations: []string{"op"}},
		"operations":     {ID: "x", Family: "f", OracleIDs: []string{"a", "b", "c"}},
	} {
		if err := validateBatch(b); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v", want, err)
		}
	}
}

func TestAddReviewedBatchRejectsInvalidPlansAtomically(t *testing.T) {
	for _, scenario := range []string{
		"hash", "owner", "batch flag", "batch count", "duplicate batch", "review", "size", "duplicate member",
		"duplicate record", "unknown", "baseline", "printing", "name", "assigned", "blocked", "excluded",
		"installed-compiled", "implemented-verified", "extra record", "reassessment", "bad prerequisite", "bad evidence",
	} {
		t.Run(scenario, func(t *testing.T) {
			q, plan, _, _ := continuationFixture(t)
			owner, batchID := "writer", "third"
			recordIndex := slices.IndexFunc(q.Records, func(r record) bool { return r.OracleID == plan.Records[2].OracleID })
			switch scenario {
			case "hash":
				plan.CorpusSHA256 = "stale"
			case "owner":
				owner = ""
			case "batch flag":
				batchID = "other"
			case "batch count":
				plan.Batches = append(plan.Batches, plan.Batches[0])
			case "duplicate batch":
				plan.Batches[0].ID, batchID = "first", "first"
			case "review":
				plan.Batches[0].Review = " "
			case "size":
				plan.Batches[0].OracleIDs = plan.Batches[0].OracleIDs[:2]
				plan.Records = plan.Records[:2]
			case "duplicate member":
				plan.Batches[0].OracleIDs[2] = plan.Batches[0].OracleIDs[0]
			case "duplicate record":
				plan.Records[2] = plan.Records[0]
			case "unknown", "baseline":
				id := "missing"
				if scenario == "baseline" {
					id = "existing"
				}
				plan.Batches[0].OracleIDs[2] = id
				plan.Records[2] = reviewedRecord{OracleID: id, Revision: "printing-" + id, Name: id}
			case "printing":
				plan.Records[2].Revision = "stale"
			case "name":
				plan.Records[2].Name = "wrong name"
			case "assigned":
				q.Records[recordIndex].Batch = "historical"
			case "blocked", "excluded", "installed-compiled", "implemented-verified":
				q.Records[recordIndex].Status = scenario
			case "extra record":
				plan.Records = append(plan.Records, plan.Records[0])
			case "reassessment":
				plan.Records[2].Reassessment = &prerequisiteEvidence{Issue: 3274}
			case "bad prerequisite":
				plan.Batches[0].Prerequisites = []int{0}
			case "bad evidence":
				plan.Batches[0].PrerequisiteEvidence = []prerequisiteEvidence{{Issue: 3274, Evidence: "issue closed"}}
			default:
				t.Fatal("unknown scenario")
			}
			before := snapshot(t, q)
			aliased := q.Records
			if err := continueQueue(&q, "add-reviewed-batch", batchID, owner, plan); err == nil {
				t.Fatal("invalid plan accepted")
			}
			if snapshot(t, q) != before || snapshot(t, aliased) != snapshot(t, q.Records) {
				t.Fatal("rejected plan partially changed in-memory state")
			}
		})
	}
}

func TestPrerequisiteEvidenceCannotRewriteHistory(t *testing.T) {
	for _, scenario := range []string{"family", "members", "prerequisites", "operations", "review", "records", "duplicate", "blank", "unknown"} {
		t.Run(scenario, func(t *testing.T) {
			q, plan, _, _ := continuationFixture(t)
			if err := continueQueue(&q, "add-reviewed-batch", "third", "writer", plan); err != nil {
				t.Fatal(err)
			}
			plan.Records = nil
			plan.Batches[0].PrerequisiteEvidence = []prerequisiteEvidence{{Issue: 3274, RuntimeRevision: "commit", Evidence: "actual runtime/scenario"}}
			switch scenario {
			case "family":
				plan.Batches[0].Family = "rewrite"
			case "members":
				plan.Batches[0].OracleIDs = []string{"x", "y", "z"}
			case "prerequisites":
				plan.Batches[0].Prerequisites = []int{3274, 999}
			case "operations":
				plan.Batches[0].Operations = []string{"other"}
			case "review":
				plan.Batches[0].Review = "rewrite"
			case "records":
				plan.Records = []reviewedRecord{{OracleID: "d"}}
			case "duplicate":
				if err := continueQueue(&q, "record-prerequisites", "third", "writer", plan); err != nil {
					t.Fatal(err)
				}
			case "blank":
				plan.Batches[0].PrerequisiteEvidence[0].Evidence = " "
			case "unknown":
				plan.Batches[0].PrerequisiteEvidence[0].Issue = 999
			}
			before := snapshot(t, q)
			if err := continueQueue(&q, "record-prerequisites", "third", "writer", plan); err == nil || snapshot(t, q) != before {
				t.Fatal("prerequisite update rewrote history or changed state on rejection")
			}
		})
	}
}

func TestTransitionRejectionsAreAtomic(t *testing.T) {
	for _, operation := range []string{"verify", "release", "block", "unknown"} {
		t.Run(operation, func(t *testing.T) {
			q, _, inventory := queueFixture(t)
			if err := transition(&q, "claim", "first", "writer", 0, "", inventory); err != nil {
				t.Fatal(err)
			}
			// The first record passes, but later verification/ownership fails.
			inventory.Production = append(inventory.Production, cards.Support{
				OracleID: "a", Loadable: true, BehaviorVerified: true, Verification: "actual a scenario",
			})
			if operation != "verify" {
				index := slices.IndexFunc(q.Records, func(r record) bool { return r.OracleID == "b" })
				q.Records[index].Owner = "other"
			}
			before := snapshot(t, q)
			aliased := q.Records
			if err := transition(&q, operation, "first", "writer", 123, "actual missing behavior", inventory); err == nil {
				t.Fatal("mixed/partial transition accepted")
			}
			if snapshot(t, q) != before || snapshot(t, aliased) != snapshot(t, q.Records) {
				t.Fatal("failed transition changed early records")
			}
		})
	}
	q, _, inventory := queueFixture(t)
	q.Records = slices.DeleteFunc(q.Records, func(r record) bool { return r.OracleID == "c" })
	for i := range q.Records {
		if q.Records[i].Batch == "first" {
			q.Records[i].Status, q.Records[i].Owner = "claimed", "writer"
		}
	}
	before := snapshot(t, q)
	if err := transition(&q, "release", "first", "writer", 0, "", inventory); err == nil || snapshot(t, q) != before {
		t.Fatal("invalid member count changed records before rejecting")
	}
}

func TestTransitionRejectsUnknownOperationAndMissingBlockerEvidence(t *testing.T) {
	for _, scenario := range []struct {
		operation string
		blocker   int
		reason    string
	}{
		{operation: "unknown"},
		{operation: "block", reason: "missing rule"},
		{operation: "block", blocker: 999},
		{operation: "block", blocker: 999, reason: " "},
	} {
		q, _, inventory := queueFixture(t)
		if err := transition(&q, "claim", "first", "writer", 0, "", inventory); err != nil {
			t.Fatal(err)
		}
		before := snapshot(t, q)
		if err := transition(&q, scenario.operation, "first", "writer", scenario.blocker, scenario.reason, inventory); err == nil ||
			snapshot(t, q) != before {
			t.Fatal("invalid operation/blocker evidence changed a valid claim")
		}
	}
}

func TestReassessPreservesCanonicalBlockerWithoutPromotingSupport(t *testing.T) {
	q, plan, corpus, inventory := continuationFixture(t)
	q.Records[slices.IndexFunc(q.Records, func(r record) bool { return r.OracleID == "d" })].Status = "blocked"
	index := slices.IndexFunc(q.Records, func(r record) bool { return r.OracleID == "d" })
	q.Records[index].Blocker, q.Records[index].Reason = 999, "observed original missing rule"
	candidate := plan.Records[0]
	candidate.Reassessment = &prerequisiteEvidence{Issue: 999, RuntimeRevision: "fixed-runtime", Evidence: "actual fixed path and passing acceptance probe; entire card expressible"}
	plan.Batches = nil
	plan.Records = []reviewedRecord{candidate}
	if err := continueQueue(&q, "reassess", "", "writer", plan); err != nil {
		t.Fatal(err)
	}
	r := q.Records[index]
	if r.Status != "unclaimed" || r.Blocker != 999 || r.Reason != "observed original missing rule" ||
		r.Evidence != "" || len(r.Reassessments) != 1 || r.Reassessments[0].Reason != r.Reason {
		t.Fatalf("reassessment erased blocker trace or invented support: %+v", r)
	}
	var resumed manifest
	if err := json.Unmarshal([]byte(snapshot(t, q)), &resumed); err != nil || snapshot(t, resumed) != snapshot(t, q) {
		t.Fatal("reassessment evidence did not survive durable roundtrip", err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if next, err := nextBatch(q, "writer"); err != nil || next != nil {
		t.Fatal("reassessment invented batch admission")
	}
}

func TestReassessRejectsPartialOrUnevidencedReopening(t *testing.T) {
	for _, scenario := range []string{"missing evidence", "wrong blocker", "printing", "duplicate", "mixed", "partial batch"} {
		t.Run(scenario, func(t *testing.T) {
			q, plan, _, _ := continuationFixture(t)
			if err := continueQueue(&q, "add-reviewed-batch", "third", "writer", plan); err != nil {
				t.Fatal(err)
			}
			for i := range q.Records {
				if q.Records[i].Batch == "third" {
					q.Records[i].Status, q.Records[i].Blocker, q.Records[i].Reason = "blocked", 999, "observed missing rule"
				}
			}
			plan.Batches = nil
			for i := range plan.Records {
				plan.Records[i].Reassessment = &prerequisiteEvidence{Issue: 999, RuntimeRevision: "fixed", Evidence: "actual runtime/scenario"}
			}
			switch scenario {
			case "missing evidence":
				plan.Records[2].Reassessment = nil
			case "wrong blocker":
				plan.Records[2].Reassessment.Issue = 1000
			case "printing":
				plan.Records[2].Revision = "stale"
			case "duplicate":
				plan.Records[2] = plan.Records[0]
			case "mixed":
				index := slices.IndexFunc(q.Records, func(r record) bool { return r.OracleID == plan.Records[2].OracleID })
				q.Records[index].Status = "unclaimed"
			case "partial batch":
				plan.Records = plan.Records[:2]
			}
			before := snapshot(t, q)
			if err := continueQueue(&q, "reassess", "", "writer", plan); err == nil || snapshot(t, q) != before {
				t.Fatal("rejected reassessment partially reopened records")
			}
		})
	}
}

func TestLegacyBatchesRemainReadableButNeedExplicitPrerequisiteEvidence(t *testing.T) {
	q, corpus, inventory := queueFixture(t)
	for i := range q.Batches {
		q.Batches[i].PrerequisiteEvidence = nil
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal("old saved batch format is incompatible", err)
	}
	if next, err := nextBatch(q, "writer"); err != nil || next != nil {
		t.Fatal("legacy issue numbers were treated as checked readiness")
	}
	for i := range q.Records {
		if q.Records[i].Batch == "first" {
			q.Records[i].Status, q.Records[i].Owner = "claimed", "writer"
		}
	}
	if next, err := nextBatch(q, "writer"); err != nil || next == nil || next.ID != "first" {
		t.Fatal("new evidence fields stranded an existing same-owner claim")
	}
}

func TestMixedBatchRetirementPreservesHistoryAndReadySiblings(t *testing.T) {
	q, corpus, inventory := queueFixture(t)
	q.CorpusSHA256 = "pin"
	for i := range q.Records {
		if q.Records[i].Batch == "second" {
			q.Records[i].Batch = ""
		}
	}
	q.Batches = q.Batches[:1]
	original := snapshot(t, q.Batches[0])
	affected := reviewedRecord{
		OracleID: "a", Revision: "printing-a", Name: "a", Blocker: 999, Reason: "observed rule missing only for a",
	}
	retirement := reviewedPlan{CorpusSHA256: "pin", Records: []reviewedRecord{affected}}
	if err := transition(&q, "claim", "first", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, q)
	if err := continueQueue(&q, "retire-batch", "first", "writer", retirement); err == nil || snapshot(t, q) != before {
		t.Fatal("retirement bypassed the active claim")
	}
	if err := transition(&q, "release", "first", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	bad := retirement
	bad.Records = []reviewedRecord{affected, {OracleID: "missing", Revision: "stale", Name: "missing", Blocker: 999, Reason: "missing"}}
	before = snapshot(t, q)
	if err := continueQueue(&q, "retire-batch", "first", "writer", bad); err == nil || snapshot(t, q) != before {
		t.Fatal("rejected partial retirement changed membership or statuses")
	}
	if err := continueQueue(&q, "retire-batch", "first", "writer", retirement); err != nil {
		t.Fatal(err)
	}
	if len(q.Batches) != 0 || len(q.RetiredBatches) != 1 || snapshot(t, q.RetiredBatches[0]) != original {
		t.Fatal("retirement lost or rewrote original batch history")
	}
	for _, r := range q.Records {
		if r.OracleID == "a" && (r.Status != "blocked" || r.Blocker != 999) {
			t.Fatal("actual blocked card was not retained")
		}
		if (r.OracleID == "b" || r.OracleID == "c") && (r.Status != "unclaimed" || r.Batch != "" || r.Blocker != 0) {
			t.Fatal("ready sibling was falsely blocked or remained assigned")
		}
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	replacement := q.RetiredBatches[0]
	replacement.ID, replacement.Review = "replacement", "reviewed b/c/d source/runtime packet"
	replacement.OracleIDs = []string{"b", "c", "d"}
	plan := reviewedPlan{CorpusSHA256: "pin", Batches: []batch{replacement}}
	for _, id := range replacement.OracleIDs {
		plan.Records = append(plan.Records, reviewedRecord{OracleID: id, Revision: "printing-" + id, Name: id})
	}
	reused := plan
	reused.Batches = slices.Clone(plan.Batches)
	reused.Batches[0].ID = "first"
	before = snapshot(t, q)
	if err := continueQueue(&q, "add-reviewed-batch", "first", "writer", reused); err == nil || snapshot(t, q) != before {
		t.Fatal("retired batch ID was reused")
	}
	if err := continueQueue(&q, "add-reviewed-batch", "replacement", "writer", plan); err != nil {
		t.Fatal(err)
	}
	if err := transition(&q, "claim", "replacement", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if pending := pendingPrerequisites(q); len(pending) != 1 || !slices.Equal(pending[0].Issues, []int{999}) ||
		!slices.Equal(pending[0].OracleIDs, []string{"a"}) {
		t.Fatalf("retired dependent's canonical prerequisite disappeared: %+v", pending)
	}
	if snapshot(t, q.RetiredBatches[0]) != original {
		t.Fatal("new assignment rewrote retired history")
	}
}

func TestReviewedPlanRejectsUnknownFieldsAndTrailingInput(t *testing.T) {
	for _, input := range []string{
		`{"corpus_sha256":"pin","ready":true}`,
		`{"batches":[{"id":"x","ready":true}]}`,
		`{"records":[{"oracle_id":"a","status":"implemented-verified"}]}`,
		`{} {}`,
		`null`,
	} {
		t.Run(input, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "plan.json")
			if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
				t.Fatal(err)
			}
			plan, err := loadReviewedPlan(path)
			if err == nil && plan.CorpusSHA256 != "" {
				t.Fatal("invalid plan decoded as a usable mutation")
			}
			if err == nil {
				t.Fatal("unknown or trailing JSON accepted")
			}
		})
	}
}
