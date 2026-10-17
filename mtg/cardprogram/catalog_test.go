package cardprogram

import (
	"encoding/json"
	"errors"
	"slices"
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

func TestCatalogScheduledDependenciesAreVersionedAndCopied(t *testing.T) {
	t.Parallel()
	ref := game.ProgramReference{ID: "creator", Version: "1"}
	delayed := game.ProgramReference{ID: "delayed", Version: "1"}
	dependencies := []game.ProgramReference{delayed}
	parent := Program{Reference: ref, Resolve: resolveNoop, Dependencies: dependencies}
	child := Program{Reference: delayed, Resolve: resolveNoop}
	if _, err := NewCatalog([]Program{parent}); err == nil {
		t.Fatal("omitted scheduled handler was admitted")
	}

	wrong := child
	wrong.Reference.Version = "2"
	if _, err := NewCatalog([]Program{parent, wrong}); err == nil {
		t.Fatal("wrong scheduled handler version was admitted")
	}
	catalog, err := NewCatalog([]Program{parent, child})
	if err != nil {
		t.Fatal(err)
	}
	dependencies[0].Version = "changed"
	lookedUp, err := catalog.Lookup(ref)
	if err != nil || lookedUp.Dependencies[0] != delayed {
		t.Fatal("catalog aliased declaration dependencies")
	}
	lookedUp.Dependencies[0].Version = "changed"
	again, err := catalog.Lookup(ref)
	if err != nil || again.Dependencies[0] != delayed {
		t.Fatal("Lookup exposed mutable catalog dependencies")
	}
}

func TestReferencesIncludeTransitiveDependencies(t *testing.T) {
	t.Parallel()
	parent := game.ProgramReference{ID: "parent", Version: "1"}
	child := game.ProgramReference{ID: "child", Version: "2"}
	delayed := game.ProgramReference{ID: "delayed", Version: "3"}
	catalog, err := NewCatalog([]Program{
		{Reference: parent, Resolve: resolveNoop, Dependencies: []game.ProgramReference{child, delayed}},
		{Reference: child, Resolve: resolveNoop, Dependencies: []game.ProgramReference{delayed}},
		{Reference: delayed, Resolve: resolveNoop, Dependencies: []game.ProgramReference{parent}},
		{Reference: game.ProgramReference{ID: "unrelated", Version: "1"}, Resolve: resolveNoop},
	})
	if err != nil {
		t.Fatal(err)
	}
	references, err := catalog.References([]game.ProgramReference{parent, parent})
	if err != nil {
		t.Fatal(err)
	}
	want := []game.ProgramReference{child, delayed, parent}
	if !slices.Equal(references, want) {
		t.Fatalf("references = %v, want %v", references, want)
	}
	references[0].Version = "changed"
	again, err := catalog.References([]game.ProgramReference{parent})
	if err != nil || !slices.Equal(again, want) {
		t.Fatalf("references exposed mutable catalog state: %v (%v)", again, err)
	}
	for _, invalid := range []game.ProgramReference{
		{ID: "missing", Version: "1"},
		{ID: parent.ID, Version: "wrong"},
	} {
		if _, err := catalog.References([]game.ProgramReference{invalid}); err == nil {
			t.Fatalf("accepted missing/mismatched reference %v", invalid)
		}
	}
}
