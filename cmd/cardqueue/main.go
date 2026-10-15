// Command cardqueue maintains the bounded authored-card work queue, not card code.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/internal/supportinventory"
	"github.com/natefinch/council4/mtg/cards"
)

type record struct {
	OracleID string   `json:"oracle_id"`
	Revision string   `json:"revision"`
	Name     string   `json:"name"`
	Status   string   `json:"status,omitempty"`
	Family   string   `json:"family,omitempty"`
	Decks    int      `json:"deck_occurrences,omitempty"`
	Batch    string   `json:"batch,omitempty"`
	Owner    string   `json:"owner,omitempty"`
	Blocker  int      `json:"blocker,omitempty"`
	Reason   string   `json:"reason,omitempty"`
	Evidence string   `json:"evidence,omitempty"`
	Keys     []string `json:"registration_keys,omitempty"`
}

type batch struct {
	ID            string   `json:"id"`
	Family        string   `json:"family"`
	OracleIDs     []string `json:"oracle_ids"`
	Prerequisites []int    `json:"prerequisites"`
	Operations    []string `json:"operations"`
}

type manifest struct {
	Issue            int             `json:"issue"`
	Baseline         string          `json:"baseline_head"`
	CorpusSHA256     string          `json:"corpus_sha256"`
	Rules            string          `json:"rules_effective"`
	Policy           string          `json:"policy"`
	Order            string          `json:"ordering"`
	Effort           string          `json:"effort"`
	DeckSources      json.RawMessage `json:"deck_sources"`
	RulesEvidence    json.RawMessage `json:"rules_evidence"`
	Reconciliation   json.RawMessage `json:"issue_reconciliation"`
	HighestBlocker   json.RawMessage `json:"highest_leverage_blocker"`
	SelectionChanges json.RawMessage `json:"selection_changes,omitempty"`
	SharedChanges    json.RawMessage `json:"shared_prerequisite_changes,omitempty"`
	Batches          []batch         `json:"batches"`
	Records          []record        `json:"records"`
	// Baseline contains associations, not invented exact legacy registrations.
	BaselineProduction    []record `json:"baseline_production"`
	CorrectedAssociations []record `json:"corrected_name_only_associations,omitempty"`
}

func main() {
	operation := flag.String("operation", "check", "init, check, next, claim, verify, release, block, format, reconcile-compiled")
	path := flag.String("manifest", "testdata/card-authoring/production.json", "saved queue")
	plan := flag.String("plan", ".cardwork/production-plan.json", "init seed: reviewed batches, blockers and evidence")
	corpus := flag.String("corpus", ".cardwork/corpus/oracle-cards.json", "pinned Oracle corpus")
	compiler := flag.String("compiler-report", ".cardwork/preservation/compile-report.json", "frozen evidence; never compiled")
	batchID := flag.String("batch", "", "explicit batch for mutation")
	owner := flag.String("owner", "", "sole writer session identity")
	blocker := flag.Int("blocker", 0, "precise open capability issue for block")
	reason := flag.String("reason", "", "observed missing behavior for block")
	flag.Parse()
	if err := run(*operation, *path, *plan, *corpus, *compiler, *batchID, *owner, *blocker, *reason); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(operation, path, planPath, corpusPath, compilerPath, batchID, owner string, blocker int, reason string) error {
	corpusData, err := os.ReadFile(corpusPath)
	if err != nil {
		return err
	}
	var corpus []cardgen.ScryfallCard
	if err := json.Unmarshal(corpusData, &corpus); err != nil {
		return err
	}
	registry := cards.NewDefaultRegistry()
	inventory, err := supportinventory.Load(corpusPath, compilerPath, registry)
	if err != nil {
		return err
	}
	source := path
	if operation == "init" {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			return errors.New("init refuses to overwrite a saved manifest")
		}
		source = planPath
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	var queue manifest
	if err := json.Unmarshal(data, &queue); err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(corpusData))
	if queue.CorpusSHA256 != hash {
		return errors.New("corpus revision changed: reconcile explicitly, never reset claims")
	}
	if operation == "init" {
		if err := initialize(&queue, corpus, inventory); err != nil {
			return err
		}
	}
	if operation == "reconcile-compiled" {
		if err := reconcileCompiled(&queue, inventory); err != nil {
			return err
		}
	}
	if err := validate(queue, corpus, inventory); err != nil {
		return err
	}
	switch operation {
	case "check":
		return printCounts(queue)
	case "next":
		next, err := nextBatch(queue, owner)
		if err != nil {
			return err
		}
		if next == nil {
			fmt.Println("No executable saved batch. Stop: review one concrete blocker or request a bounded new batch.")
			return nil
		}
		var selected []record
		for _, oracleID := range next.OracleIDs {
			for _, r := range queue.Records {
				if r.OracleID == oracleID {
					selected = append(selected, r)
				}
			}
		}
		return json.NewEncoder(os.Stdout).Encode(struct {
			Batch *batch   `json:"batch"`
			Cards []record `json:"cards"`
		}{next, selected})
	case "init", "reconcile-compiled":
	case "format":
		slices.SortFunc(queue.Records, compareRecords)
	case "claim", "verify", "release", "block":
		if err := transition(&queue, operation, batchID, owner, blocker, reason, inventory); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown operation %q", operation)
	}
	if err := validate(queue, corpus, inventory); err != nil {
		return err
	}
	data, err = json.MarshalIndent(queue, "", " ")
	if err != nil {
		return err
	}
	// O_EXCL protects first creation; the ordinary loop requires one main writer.
	if operation == "init" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(append(data, '\n'))
		closeErr := file.Close()
		return errors.Join(writeErr, closeErr)
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func initialize(queue *manifest, corpus []cardgen.ScryfallCard, inventory supportinventory.Inventory) error {
	seed := make(map[string]record)
	for _, r := range queue.Records {
		if _, exists := seed[r.OracleID]; exists {
			return fmt.Errorf("duplicate seed %s", r.OracleID)
		}
		seed[r.OracleID] = r
	}
	queue.Records = nil
	backlog := make(map[string]bool)
	associations := make(map[string]supportinventory.Record)
	for _, r := range inventory.Backlog {
		backlog[r.OracleID] = true
	}
	for _, r := range inventory.Corpus {
		associations[r.OracleID] = r
	}
	for _, card := range corpus {
		if cardgen.DisownedCard(card) {
			continue
		}
		r := seed[card.OracleID]
		delete(seed, card.OracleID)
		r.OracleID, r.Revision, r.Name = card.OracleID, card.ID, card.Name
		exclusion, excluded := (cardgen.CorpusPolicy{}).Exclusion(card)
		switch {
		case excluded:
			r.Status, r.Reason = "excluded", string(exclusion)
		case !backlog[card.OracleID]:
			association := associations[card.OracleID]
			r.Status = ""
			r.Keys = association.LegacyCandidates
			if association.Production != nil {
				r.Keys = append(r.Keys, association.Production.Key)
			}
			queue.BaselineProduction = append(queue.BaselineProduction, r)
			continue
		default:
			if r.Status == "" {
				r.Status = "unclaimed"
			}
		}
		if r.Family == "" && !excluded {
			r.Family = "unreviewed/" + card.Layout
		}
		for _, b := range queue.Batches {
			if slices.Contains(b.OracleIDs, r.OracleID) {
				r.Batch, r.Family = b.ID, b.Family
			}
		}
		queue.Records = append(queue.Records, r)
	}
	if len(seed) != 0 {
		return fmt.Errorf("seed identities absent from pinned corpus: %v", seed)
	}
	slices.SortFunc(queue.Records, compareRecords)
	slices.SortFunc(queue.BaselineProduction, func(a, b record) int { return strings.Compare(a.OracleID, b.OracleID) })
	return nil
}

func compareRecords(a, b record) int {
	if a.Decks != b.Decks {
		return b.Decks - a.Decks
	}
	if comparison := strings.Compare(a.Family, b.Family); comparison != 0 {
		return comparison
	}
	return strings.Compare(a.OracleID, b.OracleID)
}

func nextBatch(queue manifest, owner string) (*batch, error) {
	for _, r := range queue.Records {
		if r.Status == "claimed" && r.Owner != owner {
			return nil, fmt.Errorf("active writer %s on %s: coordinate, do not take over", r.Owner, r.Batch)
		}
	}
	for _, status := range []string{"claimed", "unclaimed"} {
		for _, b := range queue.Batches {
			count := 0
			for _, r := range queue.Records {
				if r.Batch == b.ID && r.Status == status {
					count++
				}
			}
			if count == len(b.OracleIDs) {
				return &b, nil
			}
		}
	}
	return nil, nil
}

func transition(queue *manifest, operation, batchID, owner string, blocker int, reason string, inventory supportinventory.Inventory) error {
	if batchID == "" || owner == "" {
		return errors.New("mutations require explicit -batch and -owner")
	}
	if operation == "claim" {
		next, err := nextBatch(*queue, owner)
		if err != nil {
			return err
		}
		if next == nil || next.ID != batchID {
			return errors.New("claim must select the next saved executable batch")
		}
	}
	verified := make(map[string]string)
	for _, support := range inventory.Production {
		if support.OracleID != "" && support.Loadable && support.BehaviorVerified {
			verified[support.OracleID] = support.Verification
		}
	}
	count := 0
	for i := range queue.Records {
		r := &queue.Records[i]
		if r.Batch != batchID {
			continue
		}
		count++
		if operation != "claim" && (r.Status != "claimed" || r.Owner != owner) {
			return fmt.Errorf("%s is not claimed by %s", r.OracleID, owner)
		}
		switch operation {
		case "claim":
			if r.Status != "unclaimed" && (r.Status != "claimed" || r.Owner != owner) {
				return fmt.Errorf("%s is not available", r.OracleID)
			}
			r.Status, r.Owner = "claimed", owner
		case "verify":
			if verified[r.OracleID] == "" {
				return fmt.Errorf("%s has no exact loadable behavior evidence; run scenarios before promotion", r.OracleID)
			}
			r.Status, r.Owner, r.Evidence = "implemented-verified", "", verified[r.OracleID]
		case "release":
			r.Status, r.Owner = "unclaimed", ""
		case "block":
			if blocker <= 0 || reason == "" {
				return errors.New("blocking needs one precise open issue and observed missing behavior")
			}
			r.Status, r.Owner, r.Blocker, r.Reason = "blocked", "", blocker, reason
		default:
			return errors.New("invalid transition")
		}
	}
	if count < 3 || count > 5 {
		return errors.New("batch must contain 3–5 records")
	}
	return nil
}

func printCounts(queue manifest) error {
	counts := make(map[string]int)
	for _, r := range queue.Records {
		counts[r.Status]++
	}
	fmt.Printf("Baseline associations %d; ", len(queue.BaselineProduction))
	return json.NewEncoder(os.Stdout).Encode(counts)
}
