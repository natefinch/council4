package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/natefinch/council4/internal/supportinventory"
)

type oracleMetadata struct {
	OracleID   string `json:"oracle_id"`
	Revision   string `json:"id"`
	Name       string `json:"name"`
	Text       string `json:"oracle_text"`
	RulingsURI string `json:"rulings_uri"`
}

func verifyBatch(queue *manifest, batchID, owner string, inventory supportinventory.Inventory, corpusData []byte) error {
	if fmt.Sprintf("%x", sha256.Sum256(corpusData)) != queue.CorpusSHA256 {
		return errors.New("verification metadata must come from the saved pinned corpus")
	}
	updated := *queue
	if err := transition(&updated, "verify", batchID, owner, 0, "", inventory); err != nil {
		return err
	}
	var corpus []oracleMetadata
	if err := json.Unmarshal(corpusData, &corpus); err != nil {
		return err
	}
	byID := make(map[string]oracleMetadata)
	for _, card := range corpus {
		if _, exists := byID[card.OracleID]; exists || card.OracleID == "" {
			return fmt.Errorf("ambiguous verification metadata %s", card.OracleID)
		}
		byID[card.OracleID] = card
	}
	rules := make(map[string]json.RawMessage)
	if len(queue.RulesEvidence) != 0 {
		if err := json.Unmarshal(queue.RulesEvidence, &rules); err != nil {
			return err
		}
		if rules == nil {
			return errors.New("rules_evidence must be an object")
		}
	}
	var sources []map[string]json.RawMessage
	if data, exists := rules["cards"]; exists {
		if err := json.Unmarshal(data, &sources); err != nil {
			return err
		}
	}
	indices := make(map[string]int)
	for i, source := range sources {
		var id string
		if err := json.Unmarshal(source["oracle_id"], &id); err != nil {
			return err
		}
		if _, exists := indices[id]; exists || id == "" {
			return fmt.Errorf("duplicate/missing Oracle source identity %s", id)
		}
		indices[id] = i
	}
	for _, r := range updated.Records {
		if r.Batch != batchID {
			continue
		}
		card, exists := byID[r.OracleID]
		if !exists || card.Revision != r.Revision || card.Name != r.Name || card.RulingsURI == "" {
			return fmt.Errorf("stale/missing pinned verification metadata %s", r.OracleID)
		}
		index, exists := indices[r.OracleID]
		if !exists {
			index = len(sources)
			sources = append(sources, make(map[string]json.RawMessage))
		}
		// Only bind corpus metadata. Retrieval, rules reads and review verdicts
		// remain the author's existing primary-source record, never inferred here.
		for field, value := range map[string]string{
			"oracle_id": r.OracleID, "revision": card.Revision, "name": card.Name,
			"oracle_text_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(card.Text))),
			"rulings_uri":        card.RulingsURI,
		} {
			if data, exists := sources[index][field]; exists {
				var saved string
				if err := json.Unmarshal(data, &saved); err != nil {
					return err
				}
				if saved != value {
					return fmt.Errorf("conflicting saved %s for %s; reconcile source evidence explicitly", field, r.OracleID)
				}
			} else {
				data, err := json.Marshal(value)
				if err != nil {
					return err
				}
				sources[index][field] = data
			}
		}
	}
	data, err := json.Marshal(sources)
	if err != nil {
		return err
	}
	rules["cards"] = data
	updated.RulesEvidence, err = json.Marshal(rules)
	if err != nil {
		return err
	}
	*queue = updated
	return nil
}
