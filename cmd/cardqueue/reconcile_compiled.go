package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/natefinch/council4/internal/supportinventory"
)

const compiledInstallationPath = "testdata/card-preservation/installed-compiler.json"

type installedCompilerIdentity struct {
	OracleID string `json:"oracle_id"`
	Name     string `json:"name"`
	Layout   string `json:"layout"`
}

func reconcileCompiled(queue *manifest, inventory supportinventory.Inventory) error {
	for _, record := range queue.Records {
		if record.Status == "claimed" {
			return errors.New("release or complete active claims before reconciling a compiler import")
		}
	}
	data, err := os.ReadFile(compiledInstallationPath)
	if err != nil {
		return err
	}
	var installation struct {
		CorpusSHA256 string                      `json:"corpus_sha256"`
		Cards        []installedCompilerIdentity `json:"cards"`
	}
	if err := json.Unmarshal(data, &installation); err != nil {
		return err
	}
	if installation.CorpusSHA256 != queue.CorpusSHA256 {
		return errors.New("installed compiler coverage does not match the queue's corpus")
	}
	installed := make(map[string]bool)
	tokens := make(map[string]bool)
	for _, card := range installation.Cards {
		if card.OracleID == "" || installed[card.OracleID] {
			return fmt.Errorf("invalid installed compiler identity %q", card.OracleID)
		}
		installed[card.OracleID] = true
		tokens[card.OracleID] = card.Layout == "token" || card.Layout == "double_faced_token"
	}
	available := make(map[string]bool)
	backlog := make(map[string]bool)
	for _, record := range inventory.Backlog {
		backlog[record.OracleID] = true
	}
	for _, record := range inventory.Corpus {
		available[record.OracleID] = record.CompilerAccepted && installed[record.OracleID] &&
			(!backlog[record.OracleID] || tokens[record.OracleID])
	}
	kept := queue.BaselineProduction[:0]
	for _, record := range queue.BaselineProduction {
		if !backlog[record.OracleID] {
			kept = append(kept, record)
			continue
		}
		// A standalone spell cannot establish support for a prepared creature
		// sharing its alternate face; ordinary cards cannot establish tokens.
		queue.CorrectedAssociations = append(queue.CorrectedAssociations, record)
		record.Status = "unclaimed"
		record.Family = "unreviewed/identity-reconciliation"
		queue.Records = append(queue.Records, record)
	}
	queue.BaselineProduction = kept
	for index := range queue.Records {
		record := &queue.Records[index]
		if record.Status == "unclaimed" && available[record.OracleID] {
			record.Status = "installed-compiled"
			record.Evidence = compiledInstallationPath
		}
	}
	return nil
}
