package cardprogram

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func resolveNoop(Context) error { return nil }

func TestCatalogIsImmutableAndVersioned(t *testing.T) {
	t.Parallel()
	reference := game.ProgramReference{ID: "test/program", Version: "1"}
	programs := []Program{{Reference: reference, Resolve: resolveNoop}}
	catalog, err := NewCatalog(programs)
	if err != nil {
		t.Fatal(err)
	}
	programs[0].Reference.Version = "2"
	programs[0].Resolve = nil
	if _, err := catalog.Lookup(reference); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []game.ProgramReference{
		{ID: "missing", Version: "1"},
		{ID: reference.ID, Version: "2"},
		{ID: reference.ID},
	} {
		if _, err := catalog.Lookup(invalid); err == nil {
			t.Fatalf("accepted missing/mismatched binding %+v", invalid)
		}
	}
	data, err := json.Marshal(reference)
	if err != nil {
		t.Fatal(err)
	}
	var restored game.ProgramReference
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Lookup(restored); err != nil {
		t.Fatal(err)
	}
	if _, err := json.Marshal(catalog); err == nil {
		t.Fatal("a Go catalog was silently serialized as data")
	}
}

func TestCatalogRejectsDuplicatesAndMalformedPrograms(t *testing.T) {
	t.Parallel()
	valid := Program{Reference: game.ProgramReference{ID: "test", Version: "1"}, Resolve: resolveNoop}
	for _, programs := range [][]Program{
		{valid, valid},
		{{Reference: valid.Reference}},
		{{Reference: game.ProgramReference{ID: "test"}, Resolve: resolveNoop}},
		{{Reference: game.ProgramReference{ID: " ", Version: "1"}, Resolve: resolveNoop}},
	} {
		if _, err := NewCatalog(programs); err == nil {
			t.Fatal("accepted an invalid catalog")
		}
	}
	catalog, err := NewCatalog([]Program{valid})
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.Run(valid.Reference, nil); err == nil {
		t.Fatal("accepted a missing context")
	}
	failing := Program{Reference: game.ProgramReference{ID: "failed", Version: "1"},
		Resolve: func(Context) error { return errors.New("operation failed") }}
	catalog, err = NewCatalog([]Program{failing})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Lookup(failing.Reference); err != nil {
		t.Fatal(err)
	}
}
