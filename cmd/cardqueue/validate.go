package main

import (
	"errors"
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/internal/supportinventory"
)

func validate(queue manifest, corpus []cardgen.ScryfallCard, inventory supportinventory.Inventory) error {
	byID := make(map[string]cardgen.ScryfallCard)
	for _, card := range corpus {
		if cardgen.DisownedCard(card) {
			continue
		}
		if _, exists := byID[card.OracleID]; exists || card.ID == "" || card.OracleID == "" {
			return fmt.Errorf("ambiguous corpus identity %s", card.OracleID)
		}
		byID[card.OracleID] = card
	}
	backlog := make(map[string]bool)
	compiled := make(map[string]bool)
	verified := make(map[string]bool)
	loadable := make(map[string]bool)
	associations := make(map[string][]string)
	for _, r := range inventory.Backlog {
		backlog[r.OracleID] = true
	}
	for _, r := range inventory.Corpus {
		compiled[r.OracleID] = r.CompilerAccepted
		associations[r.OracleID] = slices.Clone(r.LegacyCandidates)
		if r.Production != nil {
			associations[r.OracleID] = append(associations[r.OracleID], r.Production.Key)
		}
	}
	for _, support := range inventory.Production {
		loadable[support.Key] = support.Loadable
		if support.OracleID != "" && support.Loadable && support.BehaviorVerified {
			verified[support.OracleID] = true
		}
	}
	batches := make(map[string]batch)
	members := make(map[string]string)
	for _, b := range queue.Batches {
		if _, exists := batches[b.ID]; exists || b.ID == "" || b.Family == "" ||
			len(b.OracleIDs) < 3 || len(b.OracleIDs) > 5 || len(b.Prerequisites) == 0 || len(b.Operations) == 0 {
			return fmt.Errorf("invalid batch %s", b.ID)
		}
		batches[b.ID] = b
		for _, oracleID := range b.OracleIDs {
			if _, exists := members[oracleID]; exists {
				return fmt.Errorf("duplicate batch identity %s", oracleID)
			}
			members[oracleID] = b.ID
		}
	}
	activeBatch, activeOwner := "", ""
	for _, r := range queue.Records {
		card, exists := byID[r.OracleID]
		if !exists || r.Revision != card.ID || r.Name != card.Name {
			return fmt.Errorf("duplicate/stale/missing record %s", r.OracleID)
		}
		delete(byID, r.OracleID)
		exclusion, excluded := (cardgen.CorpusPolicy{}).Exclusion(card)
		if (r.Status == "excluded") != excluded || excluded && r.Reason != string(exclusion) {
			return fmt.Errorf("policy mismatch %s", r.OracleID)
		}
		if r.Batch != members[r.OracleID] || r.Batch != "" && r.Family != batches[r.Batch].Family {
			return fmt.Errorf("dangling batch membership %s", r.OracleID)
		}
		delete(members, r.OracleID)
		if r.Status != "claimed" && r.Owner != "" {
			return fmt.Errorf("dangling owner %s", r.OracleID)
		}
		switch r.Status {
		case "unclaimed":
			if !backlog[r.OracleID] {
				return fmt.Errorf("untracked production promotion %s; verify its saved claim", r.OracleID)
			}
		case "claimed":
			if r.Batch == "" || r.Owner == "" {
				return fmt.Errorf("untracked claim %s", r.OracleID)
			}
			if activeBatch != "" && (activeBatch != r.Batch || activeOwner != r.Owner) {
				return errors.New("multiple active batches/writers")
			}
			activeBatch, activeOwner = r.Batch, r.Owner
		case "implemented-verified":
			if !verified[r.OracleID] || r.Evidence == "" || r.Batch == "" {
				return fmt.Errorf("unverified admission %s", r.OracleID)
			}
		case "installed-compiled":
			token := card.Layout == "token" || card.Layout == "double_faced_token"
			if (backlog[r.OracleID] && !token) || !compiled[r.OracleID] || r.Evidence != compiledInstallationPath {
				return fmt.Errorf("unbacked compiler installation %s", r.OracleID)
			}
		case "blocked":
			if r.Blocker <= 0 || r.Reason == "" || !backlog[r.OracleID] {
				return fmt.Errorf("invalid blocker %s", r.OracleID)
			}
		case "excluded":
		default:
			return fmt.Errorf("invalid status %s", r.Status)
		}
	}
	if len(members) != 0 {
		return fmt.Errorf("missing batch records %v", members)
	}
	if activeBatch != "" {
		for _, r := range queue.Records {
			if r.Batch == activeBatch && r.Status != "claimed" {
				return errors.New("partially claimed batch")
			}
		}
	}
	for _, r := range queue.BaselineProduction {
		card, exists := byID[r.OracleID]
		if !exists || r.Revision != card.ID || r.Name != card.Name || r.Status != "" ||
			r.Owner != "" || r.Batch != "" || len(r.Keys) == 0 || backlog[r.OracleID] {
			return fmt.Errorf("invalid baseline association %s", r.OracleID)
		}
		if _, excluded := (cardgen.CorpusPolicy{}).Exclusion(card); excluded {
			return fmt.Errorf("excluded baseline association %s", r.OracleID)
		}
		if !slices.ContainsFunc(r.Keys, func(key string) bool { return loadable[key] }) {
			return fmt.Errorf("baseline production lost %s", r.OracleID)
		}
		if slices.ContainsFunc(r.Keys, func(key string) bool { return !slices.Contains(associations[r.OracleID], key) }) {
			return fmt.Errorf("invented legacy association %s", r.OracleID)
		}
		delete(byID, r.OracleID)
	}
	if len(byID) != 0 {
		return fmt.Errorf("%d corpus records unaccounted for", len(byID))
	}
	return nil
}
