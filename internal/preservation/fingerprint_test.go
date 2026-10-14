package preservation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestDefinitionFingerprintsDoNotDependOnPointerAddresses(t *testing.T) {
	t.Parallel()
	first := &game.CardDef{CardFace: game.CardFace{Name: "Card", Types: []types.Card{types.Land}}}
	second := &game.CardDef{CardFace: game.CardFace{Name: "Card", Types: []types.Card{types.Land}}}
	a, err := definitionFingerprint(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := definitionFingerprint(second)
	if err != nil || a != b {
		t.Fatalf("equal definitions differ: %s %s %v", a, b, err)
	}
	second.OracleText = "Changed text"
	b, err = definitionFingerprint(second)
	if err != nil || a == b {
		t.Fatalf("changed definition did not differ: %s %s %v", a, b, err)
	}
}

func TestSnapshotIncludesPrivatePayloadAndOrdersMaps(t *testing.T) {
	t.Parallel()
	snapshot := func(value reflect.Value) string {
		t.Helper()
		var out strings.Builder
		if err := snapshotValue(&out, value, make(map[uintptr]bool)); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}
	if a, b := snapshot(reflect.ValueOf(map[string]int{"a": 1, "b": 2})),
		snapshot(reflect.ValueOf(map[string]int{"b": 2, "a": 1})); a != b {
		t.Fatal("map order changed fingerprint")
	}
	if a, b := snapshot(reflect.ValueOf(struct{ private int }{1})),
		snapshot(reflect.ValueOf(struct{ private int }{2})); a == b {
		t.Fatal("private payload was omitted")
	}
}

func TestSnapshotRejectsFunctionsAndCycles(t *testing.T) {
	t.Parallel()
	type node struct{ next *node }
	cycle := &node{}
	cycle.next = cycle
	for _, value := range []reflect.Value{reflect.ValueOf(func() {}), reflect.ValueOf(cycle)} {
		var out strings.Builder
		if err := snapshotValue(&out, value, make(map[uintptr]bool)); err == nil {
			t.Fatal("unsupported/cyclic state was silently fingerprinted")
		}
	}
}
