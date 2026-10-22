package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/internal/supportinventory"
	"github.com/natefinch/council4/mtg/cards"
)

func verificationFixture(t *testing.T) (manifest, supportinventory.Inventory, []byte) {
	t.Helper()
	q, corpus, inventory := queueFixture(t)
	if err := transition(&q, "claim", "first", "writer", 0, "", inventory); err != nil {
		t.Fatal(err)
	}
	var metadata []oracleMetadata
	for _, card := range corpus {
		text := "Draw a card.\n" + card.Name
		if card.OracleID == "b" {
			text = ""
		}
		metadata = append(metadata, oracleMetadata{
			OracleID: card.OracleID, Revision: card.ID, Name: card.Name,
			Text: text, RulingsURI: "https://api.scryfall.com/cards/" + card.ID + "/rulings",
		})
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	q.CorpusSHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
	q.RulesEvidence = json.RawMessage(`{
		"rules_sha256":"preserved-rules-pin",
		"comprehensive_rules":["117.2e","608.2"],
		"custom":{"nested":["keep",1,true]},
		"cards":[
			{"oracle_id":"a","revision":"printing-a","retrieved":"2026-10-22","response_sha256":"actual-response","ruling_dates":[],"rules_read":{"edition":"2026-09-25","sections":["608.2"]}},
			{"oracle_id":"unrelated","custom_source":{"keep":"verbatim values"}}
		]
	}`)
	for _, id := range []string{"a", "b", "c"} {
		inventory.Production = append(inventory.Production, cards.Support{
			OracleID: id, Loadable: true, BehaviorVerified: true, Verification: "mtg/rules/TestActual" + id,
		})
	}
	return q, inventory, data
}

func decodeRules(t *testing.T, data []byte) (map[string]json.RawMessage, []map[string]any) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	var sources []map[string]any
	if err := json.Unmarshal(fields["cards"], &sources); err != nil {
		t.Fatal(err)
	}
	return fields, sources
}

func TestVerifyBindsPinnedMetadataAndPreservesSourceRecords(t *testing.T) {
	q, inventory, corpus := verificationFixture(t)
	beforeFields, beforeSources := decodeRules(t, q.RulesEvidence)
	if err := verifyBatch(&q, "first", "writer", inventory, corpus); err != nil {
		t.Fatal(err)
	}
	fields, sources := decodeRules(t, q.RulesEvidence)
	for key, before := range beforeFields {
		if key == "cards" {
			continue
		}
		var original, current any
		if err := json.Unmarshal(before, &original); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(fields[key], &current); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(original, current) {
			t.Fatalf("unrelated rules evidence changed: %s", key)
		}
	}
	if len(sources) != 4 || !reflect.DeepEqual(sources[1], beforeSources[1]) {
		t.Fatal("unrelated full source record was lost or duplicate records were appended")
	}
	for key, before := range beforeSources[0] {
		if !reflect.DeepEqual(sources[0][key], before) {
			t.Fatalf("existing primary-source evidence changed: %s", key)
		}
	}
	var metadata []oracleMetadata
	if err := json.Unmarshal(corpus, &metadata); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b", "c"} {
		index := slices.IndexFunc(sources, func(source map[string]any) bool { return source["oracle_id"] == id })
		source := sources[index]
		card := metadata[slices.IndexFunc(metadata, func(card oracleMetadata) bool { return card.OracleID == id })]
		if source["name"] != card.Name || source["revision"] != card.Revision || source["rulings_uri"] != card.RulingsURI ||
			source["oracle_text_sha256"] != fmt.Sprintf("%x", sha256.Sum256([]byte(card.Text))) {
			t.Fatalf("wrong pinned metadata for %s: %+v", id, source)
		}
		if id != "a" && len(source) != 5 {
			t.Fatalf("metadata binding invented source retrieval/reading claims: %+v", source)
		}
	}
	for _, r := range q.Records {
		if r.Batch == "first" && (r.Status != "implemented-verified" || r.Owner != "" || r.Evidence != "mtg/rules/TestActual"+r.OracleID) {
			t.Fatal("metadata binding changed behavioral admission semantics")
		}
	}
}

func TestVerifyMetadataDoesNotInventMissingSourceReview(t *testing.T) {
	q, inventory, corpus := verificationFixture(t)
	q.RulesEvidence = nil
	if err := verifyBatch(&q, "first", "writer", inventory, corpus); err != nil {
		t.Fatal(err)
	}
	fields, sources := decodeRules(t, q.RulesEvidence)
	if len(fields) != 1 || len(sources) != 3 {
		t.Fatal("metadata binding invented rules-review fields")
	}
	for _, source := range sources {
		if len(source) != 5 {
			t.Fatalf("metadata binding invented a source-review claim: %+v", source)
		}
	}
}

func TestVerifyMetadataFailuresPreserveWholeClaim(t *testing.T) {
	for _, scenario := range []string{"hash", "printing", "name", "missing", "duplicate", "rulings", "conflict", "duplicate source", "invalid rules", "partial behavior"} {
		t.Run(scenario, func(t *testing.T) {
			q, inventory, corpus := verificationFixture(t)
			var metadata []oracleMetadata
			if err := json.Unmarshal(corpus, &metadata); err != nil {
				t.Fatal(err)
			}

			c := slices.IndexFunc(metadata, func(card oracleMetadata) bool { return card.OracleID == "c" })
			switch scenario {
			case "hash":
				q.CorpusSHA256 = "stale"
			case "printing":
				metadata[c].Revision = "stale"
			case "name":
				metadata[c].Name = "wrong"
			case "missing":
				metadata = slices.Delete(metadata, c, c+1)
			case "duplicate":
				metadata = append(metadata, metadata[c])
			case "rulings":
				metadata[c].RulingsURI = ""
			case "conflict":
				q.RulesEvidence = json.RawMessage(`{"cards":[{"oracle_id":"c","oracle_text_sha256":"stale","retrieved":"preserve"}]}`)
			case "duplicate source":
				q.RulesEvidence = json.RawMessage(`{"cards":[{"oracle_id":"a"},{"oracle_id":"a"}]}`)
			case "invalid rules":
				q.RulesEvidence = json.RawMessage(`{"cards":"not an array"}`)
			case "partial behavior":
				inventory.Production = slices.DeleteFunc(inventory.Production, func(support cards.Support) bool { return support.OracleID == "c" })
			}
			if scenario != "hash" {
				var err error
				corpus, err = json.Marshal(metadata)
				if err != nil {
					t.Fatal(err)
				}
				q.CorpusSHA256 = fmt.Sprintf("%x", sha256.Sum256(corpus))
			}
			before, aliased := snapshot(t, q), q.Records
			if err := verifyBatch(&q, "first", "writer", inventory, corpus); err == nil {
				t.Fatal("invalid metadata or partial behavior admitted")
			}
			if snapshot(t, q) != before || snapshot(t, aliased) != snapshot(t, q.Records) {
				t.Fatal("rejected metadata binding partially verified the claim")
			}
		})
	}
}
