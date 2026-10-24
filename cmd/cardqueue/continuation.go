package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type prerequisiteEvidence struct {
	Issue           int    `json:"issue"`
	RuntimeRevision string `json:"runtime_revision"`
	Evidence        string `json:"evidence"`
}

type blockerReassessment struct {
	prerequisiteEvidence
	Reason string `json:"reason"`
}

type reviewedRecord struct {
	OracleID     string                `json:"oracle_id"`
	Revision     string                `json:"revision"`
	Name         string                `json:"name"`
	Blocker      int                   `json:"blocker,omitempty"`
	Reason       string                `json:"reason,omitempty"`
	Reassessment *prerequisiteEvidence `json:"reassessment,omitempty"`
}

type reviewedPlan struct {
	CorpusSHA256 string           `json:"corpus_sha256"`
	Batches      []batch          `json:"batches,omitempty"`
	Records      []reviewedRecord `json:"records,omitempty"`
}

type pendingPrerequisite struct {
	Batch     string   `json:"batch,omitempty"`
	OracleIDs []string `json:"blocked_oracle_ids,omitempty"`
	Issues    []int    `json:"issues"`
}

func loadReviewedPlan(path string) (reviewedPlan, error) {
	file, err := os.Open(path)
	if err != nil {
		return reviewedPlan{}, err
	}
	defer file.Close()
	var plan *reviewedPlan
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return reviewedPlan{}, err
	}
	if plan == nil {
		return reviewedPlan{}, errors.New("reviewed plan must be a JSON object")
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return reviewedPlan{}, errors.New("reviewed plan must contain exactly one JSON object")
	}
	return *plan, nil
}

// validateBatch checks a batch definition. Prerequisites may be empty when the
// batch relies only on operations already on main; each listed issue still
// needs prerequisite evidence before the batch is claimable.
func validateBatch(b batch) error {
	switch {
	case strings.TrimSpace(b.ID) == "":
		return errors.New("invalid batch: missing id")
	case strings.TrimSpace(b.Family) == "":
		return fmt.Errorf("invalid batch %s: missing family", b.ID)
	case len(b.OracleIDs) < 3 || len(b.OracleIDs) > 5:
		return fmt.Errorf("invalid batch %s: needs 3–5 oracle_ids, has %d", b.ID, len(b.OracleIDs))
	case len(b.Operations) == 0:
		return fmt.Errorf("invalid batch %s: operations must name the Engine operations the cards use", b.ID)
	}
	seen := make(map[string]bool)
	for _, id := range b.OracleIDs {
		if strings.TrimSpace(id) == "" || seen[id] {
			return fmt.Errorf("duplicate/empty batch identity %q", id)
		}
		seen[id] = true
	}
	issues := make(map[int]bool)
	for _, issue := range b.Prerequisites {
		if issue <= 0 || issues[issue] {
			return fmt.Errorf("invalid/duplicate prerequisite %d", issue)
		}
		issues[issue] = true
	}
	for _, operation := range b.Operations {
		if strings.TrimSpace(operation) == "" {
			return fmt.Errorf("blank operation in batch %s", b.ID)
		}
	}
	checked := make(map[int]bool)
	for _, evidence := range b.PrerequisiteEvidence {
		if !issues[evidence.Issue] || checked[evidence.Issue] || !validPrerequisiteEvidence(evidence) {
			return fmt.Errorf("invalid/duplicate prerequisite evidence for %d", evidence.Issue)
		}
		checked[evidence.Issue] = true
	}
	return nil
}

func validPrerequisiteEvidence(evidence prerequisiteEvidence) bool {
	return evidence.Issue > 0 && strings.TrimSpace(evidence.RuntimeRevision) != "" && strings.TrimSpace(evidence.Evidence) != ""
}

func missingPrerequisites(b batch) []int {
	var missing []int
	for _, issue := range b.Prerequisites {
		if !slices.ContainsFunc(b.PrerequisiteEvidence, func(e prerequisiteEvidence) bool {
			return e.Issue == issue && validPrerequisiteEvidence(e)
		}) {
			missing = append(missing, issue)
		}
	}
	return missing
}

func pendingPrerequisites(queue manifest) []pendingPrerequisite {
	var pending []pendingPrerequisite
	for _, b := range queue.Batches {
		if !slices.ContainsFunc(queue.Records, func(r record) bool {
			return r.Batch == b.ID && (r.Status == "unclaimed" || r.Status == "blocked")
		}) {
			continue
		}
		missing := missingPrerequisites(b)
		var blocked []string
		for _, r := range queue.Records {
			if r.Batch == b.ID && r.Status == "blocked" {
				blocked = append(blocked, r.OracleID)
				if !slices.Contains(missing, r.Blocker) {
					missing = append(missing, r.Blocker)
				}
			}
		}
		if len(missing) != 0 {
			pending = append(pending, pendingPrerequisite{Batch: b.ID, OracleIDs: blocked, Issues: missing})
		}
	}
	for _, r := range queue.Records {
		if r.Batch == "" && r.Status == "blocked" {
			pending = append(pending, pendingPrerequisite{OracleIDs: []string{r.OracleID}, Issues: []int{r.Blocker}})
		}
	}
	return pending
}

func unreviewedCandidates(queue manifest) []record {
	var candidates []record
	for _, r := range queue.Records {
		if r.Status == "unclaimed" && r.Batch == "" {
			candidates = append(candidates, r)
			if len(candidates) == 5 {
				break
			}
		}
	}
	return candidates
}

func continueQueue(queue *manifest, operation, batchID, owner string, plan reviewedPlan) error {
	if strings.TrimSpace(owner) == "" {
		return errors.New("continuation mutations require explicit -owner")
	}
	if plan.CorpusSHA256 == "" || plan.CorpusSHA256 != queue.CorpusSHA256 {
		return errors.New("reviewed plan has a stale/missing corpus hash")
	}
	for _, r := range queue.Records {
		if r.Status == "claimed" {
			return fmt.Errorf("active writer %s on %s: resume or finish the claim before replanning", r.Owner, r.Batch)
		}
	}
	if operation == "reassess" {
		if batchID != "" || len(plan.Batches) != 0 {
			return errors.New("reassess selects exact records, not -batch or new batch definitions")
		}
		return reassessRecords(queue, plan.Records)
	}
	if operation == "retire-batch" {
		return retireBatch(queue, batchID, plan)
	}
	if operation != "add-reviewed-batch" && operation != "record-prerequisites" {
		return errors.New("invalid continuation operation")
	}
	if len(plan.Batches) != 1 || batchID == "" || plan.Batches[0].ID != batchID {
		return errors.New("plan must contain exactly the explicit -batch")
	}
	b := plan.Batches[0]
	if err := validateBatch(b); err != nil {
		return err
	}
	index := slices.IndexFunc(queue.Batches, func(saved batch) bool { return saved.ID == batchID })
	if operation == "record-prerequisites" {
		if index < 0 || len(plan.Records) != 0 || len(b.PrerequisiteEvidence) == 0 {
			return errors.New("record-prerequisites needs an existing batch, evidence and no records")
		}
		saved := queue.Batches[index]
		if saved.Family != b.Family || saved.Review != b.Review || !slices.Equal(saved.OracleIDs, b.OracleIDs) ||
			!slices.Equal(saved.Prerequisites, b.Prerequisites) || !slices.Equal(saved.Operations, b.Operations) {
			return errors.New("prerequisite evidence cannot rewrite the saved batch")
		}
		for _, evidence := range b.PrerequisiteEvidence {
			if slices.ContainsFunc(saved.PrerequisiteEvidence, func(e prerequisiteEvidence) bool { return e.Issue == evidence.Issue }) {
				return fmt.Errorf("prerequisite %d already has evidence; never silently overwrite it", evidence.Issue)
			}
		}
		updated := slices.Clone(queue.Batches)
		saved.PrerequisiteEvidence = append(slices.Clone(saved.PrerequisiteEvidence), b.PrerequisiteEvidence...)
		updated[index] = saved
		queue.Batches = updated
		return nil
	}
	retired := slices.ContainsFunc(queue.RetiredBatches, func(saved batch) bool { return saved.ID == batchID })
	if index >= 0 || retired || strings.TrimSpace(b.Review) == "" || len(plan.Records) != len(b.OracleIDs) {
		return errors.New("new batch needs a unique ID, review reference and exact member records")
	}
	for _, saved := range queue.Batches {
		for _, id := range b.OracleIDs {
			if slices.Contains(saved.OracleIDs, id) {
				return fmt.Errorf("%s already belongs to saved batch %s", id, saved.ID)
			}
		}
	}
	updated := slices.Clone(queue.Records)
	seen := make(map[string]bool)
	for _, candidate := range plan.Records {
		if seen[candidate.OracleID] || !slices.Contains(b.OracleIDs, candidate.OracleID) ||
			candidate.Reassessment != nil || candidate.Blocker != 0 || candidate.Reason != "" {
			return fmt.Errorf("invalid/duplicate reviewed identity %s", candidate.OracleID)
		}
		seen[candidate.OracleID] = true
		index := slices.IndexFunc(updated, func(r record) bool { return r.OracleID == candidate.OracleID })
		if index < 0 {
			return fmt.Errorf("unknown or baseline identity %s", candidate.OracleID)
		}
		r := &updated[index]
		if r.Revision != candidate.Revision || r.Name != candidate.Name || r.Status != "unclaimed" || r.Batch != "" || r.Owner != "" {
			return fmt.Errorf("stale/non-unclaimed/already-assigned identity %s", candidate.OracleID)
		}
		r.Batch, r.Family = b.ID, b.Family
	}
	queue.Records = updated
	queue.Batches = append(slices.Clone(queue.Batches), b)
	return nil
}

func reassessRecords(queue *manifest, candidates []reviewedRecord) error {
	if len(candidates) == 0 || len(candidates) > 5 {
		return errors.New("reassess requires 1–5 exact blocked identities")
	}
	updated := slices.Clone(queue.Records)
	seen := make(map[string]bool)
	for _, candidate := range candidates {
		if seen[candidate.OracleID] || candidate.Blocker != 0 || candidate.Reason != "" {
			return fmt.Errorf("duplicate reassessment %s", candidate.OracleID)
		}
		seen[candidate.OracleID] = true
		index := slices.IndexFunc(updated, func(r record) bool { return r.OracleID == candidate.OracleID })
		if index < 0 {
			return fmt.Errorf("unknown blocked identity %s", candidate.OracleID)
		}
		r := &updated[index]
		evidence := candidate.Reassessment
		if r.Status != "blocked" || r.Owner != "" || r.Revision != candidate.Revision || r.Name != candidate.Name ||
			evidence == nil || evidence.Issue != r.Blocker || !validPrerequisiteEvidence(*evidence) {
			return fmt.Errorf("reassessment needs exact blocked identity and canonical blocker/runtime/scenario evidence: %s", candidate.OracleID)
		}
		r.Reassessments = append(slices.Clone(r.Reassessments), blockerReassessment{prerequisiteEvidence: *evidence, Reason: r.Reason})
		r.Status = "unclaimed"
	}
	for _, r := range updated {
		if !seen[r.OracleID] || r.Batch == "" {
			continue
		}
		for _, member := range updated {
			if member.Batch == r.Batch && !seen[member.OracleID] {
				return errors.New("reassess must include the entire blocked saved batch; history is not resized")
			}
		}
	}
	queue.Records = updated
	return nil
}

func retireBatch(queue *manifest, batchID string, plan reviewedPlan) error {
	index := slices.IndexFunc(queue.Batches, func(b batch) bool { return b.ID == batchID })
	if batchID == "" || index < 0 || len(plan.Batches) != 0 || len(plan.Records) == 0 || len(plan.Records) > 5 {
		return errors.New("retire-batch needs an existing released batch and exact newly blocked records, not batch definitions")
	}
	b := queue.Batches[index]
	updated := slices.Clone(queue.Records)
	for i := range updated {
		r := &updated[i]
		if r.Batch == batchID {
			if r.Status != "unclaimed" || r.Owner != "" {
				return errors.New("retire-batch requires all original members released and unclaimed")
			}
			r.Batch = ""
		}
	}
	seen := make(map[string]bool)
	for _, candidate := range plan.Records {
		if seen[candidate.OracleID] || !slices.Contains(b.OracleIDs, candidate.OracleID) ||
			candidate.Blocker <= 0 || strings.TrimSpace(candidate.Reason) == "" || candidate.Reassessment != nil {
			return fmt.Errorf("retirement needs exact affected members and canonical blocker/reason: %s", candidate.OracleID)
		}
		seen[candidate.OracleID] = true
		index := slices.IndexFunc(updated, func(r record) bool { return r.OracleID == candidate.OracleID })
		if index < 0 || updated[index].Revision != candidate.Revision || updated[index].Name != candidate.Name {
			return fmt.Errorf("stale/missing retirement identity %s", candidate.OracleID)
		}
		r := &updated[index]
		r.Status, r.Blocker, r.Reason = "blocked", candidate.Blocker, candidate.Reason
	}
	queue.Records = updated
	queue.RetiredBatches = append(slices.Clone(queue.RetiredBatches), b)
	queue.Batches = slices.Delete(slices.Clone(queue.Batches), index, index+1)
	return nil
}
