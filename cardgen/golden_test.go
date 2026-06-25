package cardgen

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// updateGolden rewrites the golden snapshot files under testdata/golden when
// set. Run: go test ./cardgen/ -run TestGolden -update
//
// The harness snapshots, for a small curated set of representative cards, both
// (a) the compiler's semantic IR (the []compiler.CompiledAbility produced for
// each face) and (b) the generated executable Go source. Reviewing a new or
// changed card becomes accepting a readable diff; an unintended regression in
// the parser, compiler, lowering, or renderer surfaces as a diff across the
// suite. See README.md ("Golden snapshot harness") for the workflow.
var updateGolden = flag.Bool("update", false, "update the golden snapshot files in testdata/golden")

// goldenCard is one curated entry in the snapshot suite. Rationale documents
// which effect family the card pins so the set stays representative rather than
// redundant.
type goldenCard struct {
	// Key is the stable, filename-safe slug for this card's golden files.
	Key string
	// Rationale explains which effect family this card covers.
	Rationale string
	Card      *ScryfallCard
}

// goldenCards is the curated, deliberately small suite. Each card is a real,
// currently-supported card chosen to cover a distinct major effect family so
// the harness stays fast while still exercising the breadth of the pipeline.
// Every card must lower to non-empty source with zero diagnostics (asserted by
// TestGoldenSnapshotsAreSupported); add a card by appending an entry and
// running with -update.
var goldenCards = []goldenCard{
	{
		Key:       "lightning_bolt",
		Rationale: "single-target burn: a one-shot damage effect to any target",
		Card:      &ScryfallCard{Name: "Lightning Bolt", Layout: "normal", TypeLine: "Instant", OracleText: "Lightning Bolt deals 3 damage to any target."},
	},
	{
		Key:       "shock",
		Rationale: "single-target burn variant: fixed self-named damage, near-miss of Lightning Bolt",
		Card:      &ScryfallCard{Name: "Shock", Layout: "normal", TypeLine: "Instant", OracleText: "Shock deals 2 damage to any target."},
	},
	{
		Key:       "brainstorm",
		Rationale: "ordered multi-clause sequence: draw then choose-from-hand library reorder",
		Card:      &ScryfallCard{Name: "Brainstorm", Layout: "normal", TypeLine: "Instant", OracleText: "Draw three cards, then put two cards from your hand on top of your library in any order."},
	},
	{
		Key:       "divination",
		Rationale: "plain card draw: the simplest single-effect spell body",
		Card:      &ScryfallCard{Name: "Divination", Layout: "normal", TypeLine: "Sorcery", OracleText: "Draw two cards."},
	},
	{
		Key:       "counterspell",
		Rationale: "stack interaction: counter target spell",
		Card:      &ScryfallCard{Name: "Counterspell", Layout: "normal", TypeLine: "Instant", OracleText: "Counter target spell."},
	},
	{
		Key:       "murder",
		Rationale: "targeted removal: destroy a single creature",
		Card:      &ScryfallCard{Name: "Murder", Layout: "normal", TypeLine: "Instant", OracleText: "Destroy target creature."},
	},
	{
		Key:       "naturalize",
		Rationale: "targeted removal with a disjunctive type filter: destroy artifact or enchantment",
		Card:      &ScryfallCard{Name: "Naturalize", Layout: "normal", TypeLine: "Instant", OracleText: "Destroy target artifact or enchantment."},
	},
	{
		Key:       "giant_growth",
		Rationale: "temporary pump: +X/+X to a target until end of turn",
		Card:      &ScryfallCard{Name: "Giant Growth", Layout: "normal", TypeLine: "Instant", OracleText: "Target creature gets +3/+3 until end of turn."},
	},
	{
		Key:       "raise_the_alarm",
		Rationale: "token creation: create a fixed number of typed creature tokens",
		Card:      &ScryfallCard{Name: "Raise the Alarm", Layout: "normal", TypeLine: "Instant", OracleText: "Create two 1/1 white Soldier creature tokens."},
	},
	{
		Key:       "grizzly_bears",
		Rationale: "vanilla creature: power/toughness only, no abilities",
		Card:      &ScryfallCard{Name: "Grizzly Bears", Layout: "normal", TypeLine: "Creature — Bear", OracleText: "", Power: new("2"), Toughness: new("2")},
	},
	{
		Key:       "wall_of_omens",
		Rationale: "enters-the-battlefield triggered ability: draw a card on ETB",
		Card:      &ScryfallCard{Name: "Wall of Omens", Layout: "normal", TypeLine: "Creature — Wall", OracleText: "When this creature enters, draw a card.", Power: new("0"), Toughness: new("4")},
	},
	{
		Key:       "prodigal_sorcerer",
		Rationale: "activated tap ability: {T} cost producing targeted damage",
		Card:      &ScryfallCard{Name: "Prodigal Sorcerer", Layout: "normal", TypeLine: "Creature — Human Wizard", OracleText: "{T}: Prodigal Sorcerer deals 1 damage to any target.", Power: new("1"), Toughness: new("1")},
	},
	{
		Key:       "llanowar_elves",
		Rationale: "mana ability: {T} cost adding one mana",
		Card:      &ScryfallCard{Name: "Llanowar Elves", Layout: "normal", TypeLine: "Creature — Elf Druid", OracleText: "{T}: Add {G}.", Power: new("1"), Toughness: new("1")},
	},
	{
		Key:       "glorious_anthem",
		Rationale: "static anthem: a continuous +1/+1 buff to controlled creatures",
		Card:      &ScryfallCard{Name: "Glorious Anthem", Layout: "normal", TypeLine: "Enchantment", OracleText: "Creatures you control get +1/+1."},
	},
}

// TestGoldenSnapshotsAreSupported guards the curated set's invariant: every
// snapshot card must lower to non-empty executable source with zero
// diagnostics. A card that stops lowering would otherwise silently snapshot an
// empty source golden, hiding the regression.
func TestGoldenSnapshotsAreSupported(t *testing.T) {
	t.Parallel()
	for _, gc := range goldenCards {
		t.Run(gc.Key, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(gc.Card, "g")
			if err != nil {
				t.Fatalf("%s (%s): %v", gc.Card.Name, gc.Rationale, err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("%s (%s) produced diagnostics: %#v", gc.Card.Name, gc.Rationale, diagnostics)
			}
			if strings.TrimSpace(source) == "" {
				t.Fatalf("%s (%s) produced no source", gc.Card.Name, gc.Rationale)
			}
		})
	}
}

// TestGoldenSemanticIR snapshots the compiler's semantic IR for every curated
// card. The IR is rendered through a deterministic, zero-omitting dumper so the
// golden stays compact and a diff highlights only the fields that changed.
func TestGoldenSemanticIR(t *testing.T) {
	for _, gc := range goldenCards {
		t.Run(gc.Key, func(t *testing.T) {
			got := goldenIRSnapshot(t, gc)
			checkGolden(t, gc.Key+".ir.txt", got)
		})
	}
}

// TestGoldenGeneratedSource snapshots the rendered executable Go source for
// every curated card.
func TestGoldenGeneratedSource(t *testing.T) {
	for _, gc := range goldenCards {
		t.Run(gc.Key, func(t *testing.T) {
			source, diagnostics, err := GenerateExecutableCardSource(gc.Card, "g")
			if err != nil {
				t.Fatalf("generate %s: %v", gc.Card.Name, err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("generate %s produced diagnostics: %#v", gc.Card.Name, diagnostics)
			}
			checkGolden(t, gc.Key+".source.txt", []byte(source))
		})
	}
}

// TestGoldenSnapshotsAreDeterministic guards against unstable ordering: both
// snapshots must be byte-identical across two independent generations.
func TestGoldenSnapshotsAreDeterministic(t *testing.T) {
	t.Parallel()
	for _, gc := range goldenCards {
		t.Run(gc.Key, func(t *testing.T) {
			t.Parallel()
			firstIR := goldenIRSnapshot(t, gc)
			secondIR := goldenIRSnapshot(t, gc)
			if !bytes.Equal(firstIR, secondIR) {
				t.Errorf("%s IR snapshot is not deterministic across two generations", gc.Key)
			}
			firstSrc, _, err := GenerateExecutableCardSource(gc.Card, "g")
			if err != nil {
				t.Fatalf("generate %s: %v", gc.Card.Name, err)
			}
			secondSrc, _, err := GenerateExecutableCardSource(gc.Card, "g")
			if err != nil {
				t.Fatalf("generate %s: %v", gc.Card.Name, err)
			}
			if firstSrc != secondSrc {
				t.Errorf("%s source snapshot is not deterministic across two generations", gc.Key)
			}
		})
	}
}

// goldenIRSnapshot compiles each executable face of a curated card and renders
// the resulting semantic IR through the deterministic dumper. It mirrors the
// production parse+compile path via the shared compileFaceDocument helper so
// the snapshot tracks exactly what lowering consumes.
func goldenIRSnapshot(t *testing.T, gc goldenCard) []byte {
	t.Helper()
	var out bytes.Buffer
	faces := executableFaces(gc.Card)
	for i := range faces {
		face := faces[i]
		parsedType := ParseTypeLine(face.TypeLine)
		if len(parsedType.Types) == 0 || face.OracleText == "" {
			continue
		}
		compilation, _ := compileFaceDocument(face, parsedType)
		if len(faces) > 1 {
			_, _ = fmt.Fprintf(&out, "=== face: %s ===\n", face.Name)
		}
		dumpValue(&out, reflect.ValueOf(compilation.Abilities), 0)
	}
	return out.Bytes()
}

func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name)
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatalf("update golden %s: %v", name, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s (run with -update to create): %v", name, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("golden %s drift; run `go test ./cardgen/ -run TestGolden -update` to accept.\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

var (
	spanType     = reflect.TypeFor[shared.Span]()
	positionType = reflect.TypeFor[shared.Position]()
	stringerType = reflect.TypeFor[fmt.Stringer]()
)

// dumpValue renders a value as a deterministic, readable, zero-omitting tree.
//
// Determinism: struct fields are emitted in declaration order and slices in
// index order. The compiler IR contains no maps (whose iteration order is
// random), so the output is stable across runs.
//
// Readability: zero-valued fields (false bools, 0 numbers, empty strings, nil
// pointers/slices, all-zero structs) are omitted so only meaningful fields
// appear. Named integer enums that implement fmt.Stringer are rendered by name.
// shared.Span / shared.Position fields carry source byte offsets, not
// semantics, and are skipped to keep the snapshot focused and low-noise.
func dumpValue(out *bytes.Buffer, v reflect.Value, depth int) {
	indent := strings.Repeat("  ", depth)
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			_, _ = out.WriteString("<nil>\n")
			return
		}
		dumpValue(out, v.Elem(), depth)
	case reflect.Struct:
		if name, ok := stringerName(v); ok {
			_, _ = fmt.Fprintf(out, "%s\n", name)
			return
		}
		_, _ = out.WriteString("\n")
		t := v.Type()
		for i := range v.NumField() {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			fv := v.Field(i)
			if isSkippedType(field.Type) {
				continue
			}
			if fv.IsZero() {
				continue
			}
			_, _ = fmt.Fprintf(out, "%s  %s: ", indent, field.Name)
			dumpValue(out, fv, depth+1)
		}
	case reflect.Slice, reflect.Array:
		_, _ = out.WriteString("\n")
		for i := range v.Len() {
			_, _ = fmt.Fprintf(out, "%s  [%d]: ", indent, i)
			dumpValue(out, v.Index(i), depth+1)
		}
	default:
		if name, ok := stringerName(v); ok {
			_, _ = fmt.Fprintf(out, "%s\n", name)
			return
		}
		_, _ = fmt.Fprintf(out, "%v\n", v.Interface())
	}
}

// stringerName returns the fmt.Stringer rendering of a named scalar enum so the
// snapshot reads "spell" instead of "1". It deliberately does not apply to
// struct or pointer kinds, whose fields must still be dumped individually.
func stringerName(v reflect.Value) (string, bool) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Type().Implements(stringerType) {
			if s, ok := v.Interface().(fmt.Stringer); ok {
				return s.String(), true
			}
		}
	default:
		return "", false
	}
	return "", false
}

func isSkippedType(t reflect.Type) bool {
	return t == spanType || t == positionType
}
