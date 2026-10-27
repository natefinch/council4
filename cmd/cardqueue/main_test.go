package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/internal/supportinventory"
	"github.com/natefinch/council4/mtg/cards"
)

func queueFixture(t *testing.T) (manifest, []carddata.ScryfallCard, supportinventory.Inventory) {
	t.Helper()
	q := manifest{Batches: []batch{
		{ID: "first", Family: "destruction", OracleIDs: []string{"a", "b", "c"}, Prerequisites: []int{3273}, Operations: []string{"Context.Destroy"}},
		{ID: "second", Family: "search", OracleIDs: []string{"d", "e", "f"}, Prerequisites: []int{3274}, Operations: []string{"Context.Search"}},
	}}
	for i := range q.Batches {
		q.Batches[i].PrerequisiteEvidence = []prerequisiteEvidence{{
			Issue: q.Batches[i].Prerequisites[0], RuntimeRevision: "reviewed-runtime",
			Evidence: "complete Engine path and executed behavioral scenario",
		}}
	}
	var corpus []carddata.ScryfallCard
	var inventory supportinventory.Inventory
	for _, oracleID := range []string{"a", "b", "c", "d", "e", "f", "existing"} {
		card := carddata.ScryfallCard{
			ID: "printing-" + oracleID, OracleID: oracleID, Name: oracleID,
			Layout: "normal", Games: []string{"paper"}, Legalities: map[string]string{"legacy": "legal"},
		}
		corpus = append(corpus, card)
		r := supportinventory.Record{OracleID: oracleID, Name: oracleID}
		if oracleID == "existing" {
			r.LegacyCandidates = []string{"legacy-name:existing/0"}
			inventory.Production = append(inventory.Production, cards.Support{Key: r.LegacyCandidates[0], Name: oracleID, Loadable: true, Identity: "name-only"})
		} else {
			inventory.Backlog = append(inventory.Backlog, r)
		}
		inventory.Corpus = append(inventory.Corpus, r)
	}
	corpus = append(corpus, carddata.ScryfallCard{ID: "printing-excluded", OracleID: "excluded", Name: "excluded", Layout: "normal", Digital: true})
	if err := initialize(&q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	return q, corpus, inventory
}

func TestQueueAccountingAndClaims(t *testing.T) {
	q, corpus, inventory := queueFixture(t)
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if len(q.Records) != 7 || len(q.BaselineProduction) != 1 ||
		q.BaselineProduction[0].Keys[0] != "legacy-name:existing/0" {
		t.Fatal("remaining/excluded/legacy accounting lost a record or invented exact support")
	}
	if err := transition(&q, "claim", "first", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if _, err := nextBatch(q, "other-writer"); err == nil {
		t.Fatal("another writer took over an active claim")
	}
	if next, err := nextBatch(q, "writer"); err != nil || next.ID != "first" {
		t.Fatal("saved claim was not resumed")
	}
	if err := transition(&q, "claim", "second", "writer", 0, "", inventory); err == nil {
		t.Fatal("same writer opened a concurrent batch")
	}
	for _, oracleID := range q.Batches[0].OracleIDs {
		inventory.Production = append(inventory.Production, cards.Support{
			Key: oracleID + "/front", OracleID: oracleID, Loadable: true, BehaviorVerified: true, Verification: "actual named scenario",
		})
	}
	inventory.Backlog = inventory.Backlog[3:]
	if err := transition(&q, "verify", "first", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if next, err := nextBatch(q, ""); err != nil || next.ID != "second" {
		t.Fatal("completed batch did not advance saved queue")
	}
	if err := transition(&q, "claim", "second", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	if err := transition(&q, "verify", "second", "writer", 0, "", inventory); err == nil {
		t.Fatal("claimed drafts counted as verified production")
	}
	if err := transition(&q, "block", "second", "writer", 170, "observed missing payment", inventory); err != nil {
		t.Fatal(err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if next, err := nextBatch(q, ""); err != nil || next != nil {
		t.Fatal("blocked queue did not stop")
	}
}

func TestQueueRejectsUnaccountedWork(t *testing.T) {
	for _, scenario := range []string{"duplicate", "missing", "revision", "status", "owner", "batch", "baseline", "policy", "dangling batch", "partial claim"} {
		t.Run(scenario, func(t *testing.T) {
			q, corpus, inventory := queueFixture(t)
			first := slices.IndexFunc(q.Records, func(r record) bool { return r.Batch == "first" })
			switch scenario {
			case "duplicate":
				q.Records = append(q.Records, q.Records[0])
			case "missing":
				q.Records = q.Records[1:]
			case "revision":
				q.Records[0].Revision = "stale"
			case "status":
				q.Records[0].Status = "generated"
			case "owner":
				q.Records[0].Owner = "dangling"
			case "batch":
				q.Records[0].Batch = "unknown"
			case "baseline":
				q.BaselineProduction[0].Keys = []string{"invented"}
			case "policy":
				excluded := slices.IndexFunc(q.Records, func(r record) bool { return r.Status == "excluded" })
				q.Records[excluded].Status = "unclaimed"
			case "dangling batch":
				q.Batches[0].OracleIDs[0] = "missing"
			case "partial claim":
				q.Records[first].Status, q.Records[first].Owner = "claimed", "writer"
			default:
				t.Fatal("unknown rejection scenario")
			}
			if err := validate(q, corpus, inventory); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
}

func TestQueueIssueScopedAdmissionWithoutChangingSavedBatches(t *testing.T) {
	for _, scenario := range []string{"verified", "unloadable", "unverified", "no evidence"} {
		t.Run(scenario, func(t *testing.T) {
			q, corpus, inventory := queueFixture(t)
			q.Batches = nil
			for index := range q.Records {
				q.Records[index].Batch = ""
			}
			index := slices.IndexFunc(q.Records, func(r record) bool { return r.OracleID == "a" })
			q.Records[index].Status = "implemented-verified"
			q.Records[index].Evidence = "mtg/rules named issue scenario"
			if scenario == "no evidence" {
				q.Records[index].Evidence = ""
			}
			inventory.Backlog = slices.DeleteFunc(inventory.Backlog, func(r supportinventory.Record) bool {
				return r.OracleID == "a"
			})
			inventory.Production = append(inventory.Production, cards.Support{
				OracleID: "a", Loadable: scenario != "unloadable", BehaviorVerified: scenario != "unverified",
			})
			err := validate(q, corpus, inventory)
			if (err == nil) != (scenario == "verified") {
				t.Fatalf("admission validation: %v", err)
			}
		})
	}
}

func TestQueueDeterministicPriorityAndRoundtrip(t *testing.T) {
	q, corpus, inventory := queueFixture(t)
	q.Records[3].Decks = 2
	q.BaselineProduction = nil
	if err := initialize(&q, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	if q.Records[0].Decks != 2 {
		t.Fatal("recorded deck occurrence did not precede deterministic family/identity")
	}
	data, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	var decoded manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := validate(decoded, corpus, inventory); err != nil {
		t.Fatal(fmt.Errorf("roundtrip: %w", err))
	}
}
