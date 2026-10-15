package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/natefinch/council4/internal/supportinventory"
)

func TestReconcileCompilerInstallationDoesNotInventBehaviorEvidence(t *testing.T) {
	t.Chdir(t.TempDir())
	queue, corpus, inventory := queueFixture(t)
	queue.CorpusSHA256 = "pinned"
	installed := slices.IndexFunc(queue.Records, func(r record) bool { return r.OracleID == "a" })
	blocked := slices.IndexFunc(queue.Records, func(r record) bool { return r.OracleID == "b" })
	queue.Records[blocked].Status = "blocked"
	queue.Records[blocked].Blocker = 3289
	queue.Records[blocked].Reason = "proven runtime defect"
	for index := range inventory.Corpus {
		inventory.Corpus[index].CompilerAccepted = true
	}
	inventory.Backlog = slices.DeleteFunc(inventory.Backlog, func(r supportinventory.Record) bool { return r.OracleID == "a" })
	data := []byte(`{"corpus_sha256":"pinned","cards":[{"oracle_id":"a"},{"oracle_id":"b"}]}`)
	if err := os.MkdirAll(filepath.Dir(compiledInstallationPath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(compiledInstallationPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := reconcileCompiled(&queue, inventory); err != nil {
		t.Fatal(err)
	}
	if queue.Records[installed].Status != "installed-compiled" ||
		queue.Records[installed].Evidence != compiledInstallationPath ||
		queue.Records[blocked].Status != "blocked" {
		t.Fatal("compiler import lost its distinct status or overwrote a runtime blocker")
	}
	if err := validate(queue, corpus, inventory); err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(queue)
	if err != nil {
		t.Fatal(err)
	}
	if err := reconcileCompiled(&queue, inventory); err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(queue)
	if err != nil || string(before) != string(after) {
		t.Fatal("repeated reconciliation changed the saved queue", err)
	}
	queue.Records[installed].Status = "implemented-verified"
	if err := validate(queue, corpus, inventory); err == nil {
		t.Fatal("compiler-only installation was accepted as verified behavior")
	}
}
