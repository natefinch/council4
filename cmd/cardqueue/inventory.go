package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/internal/supportinventory"
	"github.com/natefinch/council4/mtg/cards"
)

func loadQueueInventory(corpusPath, compilerPath string, corpus []carddata.ScryfallCard, hash string, registry *cards.Registry) (supportinventory.Inventory, error) {
	if _, err := os.Stat(compilerPath); err == nil {
		return supportinventory.Load(corpusPath, compilerPath, registry)
	} else if !errors.Is(err, os.ErrNotExist) {
		return supportinventory.Inventory{}, err
	}
	data, err := os.ReadFile(compiledInstallationPath)
	if err != nil {
		return supportinventory.Inventory{}, fmt.Errorf("no frozen compiler report; read retained installation: %w", err)
	}
	var installation struct {
		CorpusSHA256 string                      `json:"corpus_sha256"`
		Cards        []installedCompilerIdentity `json:"cards"`
	}
	if err := json.Unmarshal(data, &installation); err != nil {
		return supportinventory.Inventory{}, err
	}
	if installation.CorpusSHA256 != hash {
		return supportinventory.Inventory{}, errors.New("retained installation does not match the pinned corpus")
	}
	return inventoryFromInstallation(corpus, hash, registry.Inventory(), installation.Cards)
}

// This joins retained installation evidence, not compiler acceptance for the
// uninstalled tail. Loadability and verification still come from the registry.
func inventoryFromInstallation(corpus []carddata.ScryfallCard, hash string, production []cards.Support, installation []installedCompilerIdentity) (supportinventory.Inventory, error) {
	out := supportinventory.Inventory{CorpusSHA256: hash, Production: production}
	installed := make(map[string]installedCompilerIdentity)
	for _, identity := range installation {
		if identity.OracleID == "" {
			return out, errors.New("empty installed identity")
		}
		if _, exists := installed[identity.OracleID]; exists {
			return out, fmt.Errorf("duplicate installed identity %s", identity.OracleID)
		}
		installed[identity.OracleID] = identity
	}
	exact := make(map[string]cards.Support)
	byName := make(map[string][]cards.Support)
	for _, support := range production {
		if support.OracleID != "" {
			if _, exists := exact[support.OracleID]; exists {
				return out, fmt.Errorf("duplicate production identity %s", support.OracleID)
			}
			exact[support.OracleID] = support
		} else {
			byName[support.Name] = append(byName[support.Name], support)
		}
	}
	seen := make(map[string]bool)
	for _, card := range corpus {
		if carddata.DisownedCard(card) {
			continue
		}
		if card.OracleID == "" || seen[card.OracleID] {
			return out, fmt.Errorf("ambiguous corpus identity %s", card.OracleID)
		}
		seen[card.OracleID] = true
		identity, wasInstalled := installed[card.OracleID]
		if wasInstalled && (identity.Name != card.Name || identity.Layout != card.Layout) {
			return out, fmt.Errorf("stale installation identity %s", card.OracleID)
		}
		delete(installed, card.OracleID)
		if _, excluded := (carddata.CorpusPolicy{}).Exclusion(card); excluded {
			if wasInstalled {
				return out, fmt.Errorf("excluded installation identity %s", card.OracleID)
			}
			continue
		}
		r := supportinventory.Record{OracleID: card.OracleID, Name: card.Name, CompilerAccepted: wasInstalled}
		if support, exists := exact[card.OracleID]; exists {
			if support.Name != card.Name {
				return out, fmt.Errorf("production identity name mismatch %s", card.OracleID)
			}
			r.Production = &support
			delete(exact, card.OracleID)
		}
		candidates := slices.Clone(byName[card.Name])
		for _, face := range card.CardFaces {
			for _, support := range byName[face.Name] {
				if !slices.ContainsFunc(candidates, func(candidate cards.Support) bool { return candidate.Key == support.Key }) {
					candidates = append(candidates, support)
				}
			}
		}
		legacyLoadable := false
		for _, support := range candidates {
			r.LegacyCandidates = append(r.LegacyCandidates, support.Key)
			if !support.Loadable || card.Layout == "token" || card.Layout == "double_faced_token" {
				continue
			}
			legacyLoadable = legacyLoadable || support.Name == card.Name ||
				len(card.CardFaces) > 1 && !slices.ContainsFunc(card.CardFaces, func(face carddata.ScryfallCardFace) bool {
					return !slices.Contains(support.Faces, face.Name)
				})
		}
		out.Corpus = append(out.Corpus, r)
		if (r.Production == nil || !r.Production.Loadable) && !legacyLoadable {
			out.Backlog = append(out.Backlog, r)
		}
	}
	if len(installed) != 0 || len(exact) != 0 {
		return out, errors.New("installation or production identities absent from eligible pinned corpus")
	}
	return out, nil
}
