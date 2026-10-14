package preservation

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Exception acknowledges one exact baseline loss for a documented correction.
type Exception struct {
	Inventory           string `json:"inventory"`
	Key                 string `json:"key"`
	CorpusSHA256        string `json:"corpus_sha256"`
	BaselineFingerprint string `json:"baseline_fingerprint"`
	Reason              string `json:"reason"`
	IssueURL            string `json:"issue_url"`
}

// Change records an admission, loss, definition change or diagnostic change.
type Change struct {
	Inventory    string   `json:"inventory"`
	Key          string   `json:"key"`
	Kind         string   `json:"kind"`
	Before       string   `json:"before,omitempty"`
	After        string   `json:"after,omitempty"`
	Refusals     []string `json:"refusals,omitempty"`
	Acknowledged bool     `json:"acknowledged,omitempty"`
}

// Compare reports every difference and rejects unacknowledged accepted-card losses.
// RegistryOnly does not claim to recompile the frozen compiler inventory.
func Compare(base, candidate Manifest, exceptions []Exception, registryOnly bool) ([]Change, error) {
	if err := base.Validate(); err != nil {
		return nil, fmt.Errorf("baseline: %w", err)
	}
	if err := candidate.Validate(); err != nil {
		return nil, fmt.Errorf("candidate: %w", err)
	}
	if base.Input != candidate.Input {
		return nil, errors.New("preservation inputs differ: update the corpus baseline explicitly, not during a code comparison")
	}
	for _, exception := range exceptions {
		if exception.Key == "" || exception.Reason == "" ||
			!strings.HasPrefix(exception.IssueURL, "https://github.com/natefinch/council4/issues/") ||
			exception.CorpusSHA256 != base.Input.JSONSHA256 || exception.BaselineFingerprint == "" ||
			(exception.Inventory != "compiler" && exception.Inventory != "registry") {
			return nil, errors.New("loss exception must identify the corpus, inventory, exact baseline fingerprint, reason and repository issue")
		}
	}
	var changes []Change
	losses := 0
	for index, kind := range []string{"compiler", "registry"} {
		if registryOnly && index == 0 {
			continue
		}
		oldRecords := base.Compiler
		newRecords := candidate.Compiler
		if index == 1 {
			oldRecords, newRecords = base.Registry, candidate.Registry
		}
		current := make(map[string]Record, len(newRecords))
		for _, record := range newRecords {
			current[record.Key] = record
		}
		for _, old := range oldRecords {
			newRecord, exists := current[old.Key]
			delete(current, old.Key)
			switch {
			case old.DefinitionSHA256 != "" && (!exists || newRecord.DefinitionSHA256 == ""):
				acknowledged := slices.ContainsFunc(exceptions, func(exception Exception) bool {
					return exception.Inventory == kind && exception.Key == old.Key &&
						exception.BaselineFingerprint == old.DefinitionSHA256
				})
				changes = append(changes, Change{
					Inventory: kind, Key: old.Key, Kind: "loss", Before: old.DefinitionSHA256,
					Refusals: newRecord.Refusals, Acknowledged: acknowledged,
				})
				if !acknowledged {
					losses++
				}
			case old.DefinitionSHA256 == "" && newRecord.DefinitionSHA256 != "":
				changes = append(changes, Change{Inventory: kind, Key: old.Key, Kind: "admission", After: newRecord.DefinitionSHA256})
			case old.DefinitionSHA256 != newRecord.DefinitionSHA256:
				changes = append(changes, Change{
					Inventory: kind, Key: old.Key, Kind: "changed", Before: old.DefinitionSHA256, After: newRecord.DefinitionSHA256,
				})
			case !slices.Equal(old.Refusals, newRecord.Refusals):
				changes = append(changes, Change{Inventory: kind, Key: old.Key, Kind: "diagnostics", Refusals: newRecord.Refusals})
			default:
			}
		}
		for key, record := range current {
			kindOfChange := "diagnostics"
			if record.DefinitionSHA256 != "" {
				kindOfChange = "admission"
			}
			changes = append(changes, Change{
				Inventory: kind, Key: key, Kind: kindOfChange, After: record.DefinitionSHA256, Refusals: record.Refusals,
			})
		}
	}
	slices.SortFunc(changes, func(a, b Change) int {
		return strings.Compare(a.Inventory+"\x00"+a.Key, b.Inventory+"\x00"+b.Key)
	})
	if losses > 0 {
		return changes, fmt.Errorf("%d unacknowledged supported-card losses", losses)
	}
	return changes, nil
}
