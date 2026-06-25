package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestLowerConditionSelectionRoutesThroughCanonicalProjector locks the
// byte-identical mapping after lowerConditionSelection was unified onto the
// shared SelectionForSelector projector (umbrella #1414). It exercises the full
// shared dimension cluster (conjunctive required types, supertype, subtype,
// color, tapped, combat, keyword) together with every condition-specific rider
// extra (AnyCounter, the named-counter count threshold, ExcludeSource, the
// power-at-least bound, and TokenOnly) and asserts the projected game.Selection
// equals the value the legacy hand-written projector produced.
func TestLowerConditionSelectionRoutesThroughCanonicalProjector(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		selection compiler.ConditionSelection
		want      game.Selection
		wantOK    bool
	}{
		{
			name:      "empty selection stays empty and valid",
			selection: compiler.ConditionSelection{},
			want:      game.Selection{},
			wantOK:    true,
		},
		{
			name: "shared cluster plus every rider extra",
			selection: compiler.ConditionSelection{
				RequiredTypes:       []types.Card{types.Creature},
				Supertypes:          []types.Super{types.Legendary},
				SubtypesAny:         []string{"Goblin"},
				ColorsAny:           []color.Color{color.Red},
				Tapped:              compiler.ConditionTriTrue,
				CombatState:         compiler.ConditionCombatStateAttacking,
				Keyword:             parser.KeywordFlying,
				TokenOnly:           true,
				ExcludeSource:       true,
				CounterKind:         counter.Charge,
				CounterKindKnown:    true,
				CounterCountAtLeast: 5,
				MatchPowerAtLeast:   true,
				PowerAtLeast:        3,
			},
			want: game.Selection{
				RequiredTypes:        []types.Card{types.Creature},
				Supertypes:           []types.Super{types.Legendary},
				SubtypesAny:          []types.Sub{types.Sub("Goblin")},
				ColorsAny:            []color.Color{color.Red},
				Tapped:               game.TriTrue,
				CombatState:          game.CombatStateAttacking,
				Keyword:              game.Flying,
				TokenOnly:            true,
				ExcludeSource:        true,
				RequiredCounter:      counter.Charge,
				RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5}),
				Power:                opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3}),
			},
			wantOK: true,
		},
		{
			name: "kind-agnostic any-counter rider",
			selection: compiler.ConditionSelection{
				RequiredTypes: []types.Card{types.Artifact},
				AnyCounter:    true,
			},
			want: game.Selection{
				RequiredTypes:   []types.Card{types.Artifact},
				MatchAnyCounter: true,
			},
			wantOK: true,
		},
		{
			name: "bare power-at-least without flag fails closed",
			selection: compiler.ConditionSelection{
				PowerAtLeast: 4,
			},
			want:   game.Selection{},
			wantOK: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := lowerConditionSelection(tc.selection)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v (selection %#v)", ok, tc.wantOK, got)
			}
			if !ok {
				return
			}
			// The projector clones slice fields into empty-but-non-nil slices
			// where the legacy projector did the same, so compare the rendered
			// Go source (the byte-identical criterion) rather than the in-memory
			// value, which would distinguish nil from an empty slice.
			if gotSrc, wantSrc := renderSelectionSource(t, got), renderSelectionSource(t, tc.want); gotSrc != wantSrc {
				t.Fatalf("selection mismatch\n got: %s\nwant: %s", gotSrc, wantSrc)
			}
		})
	}
}

func renderSelectionSource(t *testing.T, selection game.Selection) string {
	t.Helper()
	source, err := (Renderer{}).renderSelection(newRenderCtx(), selection)
	if err != nil {
		t.Fatalf("renderSelection(%#v) error: %v", selection, err)
	}
	return source
}

// TestLowerConditionPowerQualifierEndToEnd is an end-to-end guard that a corpus
// condition filter carrying a power-at-least qualifier still lowers identically
// through the unified projector. "draw a card if its power is 3 or greater"
// routes through lowerConditionSelection's MatchPowerAtLeast rider, so the gated
// instruction's ObjectMatches selection must keep Power >= 3.
func TestLowerConditionPowerQualifierEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Power Gate",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature you control enters, draw a card if its power is 3 or greater.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	var matched bool
	for _, instruction := range mode.Sequence {
		gate := instruction.Condition
		if !gate.Exists || !gate.Val.Condition.Exists {
			continue
		}
		selection := gate.Val.Condition.Val.ObjectMatches
		if !selection.Exists || !selection.Val.Power.Exists {
			continue
		}
		matched = true
		power := selection.Val.Power.Val
		want := compare.Int{Op: compare.GreaterOrEqual, Value: 3}
		if power != want {
			t.Fatalf("power selection = %#v, want %#v", power, want)
		}
	}
	if !matched {
		t.Fatalf("no instruction carried a power-qualifier ObjectMatches selection: %+v", mode.Sequence)
	}
}
