package cardgen

import (
	"reflect"
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

// TestLowerStaticSelectionDimensionParity exhaustively pins the
// StaticSelection -> game.Selection projection now that lowerStaticSelection
// routes its shared dimension cluster through the canonical
// SelectionForSelectorMasked instead of a second hand-written projector
// (umbrella #1414). Each case isolates one StaticSelection dimension (or a
// genuine per-context extra carried as a rider) and asserts the exact
// Selection field it produces, proving the unified routing reproduces the
// prior projector's accept set field-for-field and that the fail-closed
// reject branches (unknown combat/tap state) still reject.
func TestLowerStaticSelectionDimensionParity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		selection compiler.StaticSelection
		want      game.Selection
		wantOK    bool
	}{
		// Shared dimension cluster routed through the canonical projector.
		{
			name:      "required type single conjunctive",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
			wantOK:    true,
		},
		{
			name:      "required types multiple all-of",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
			want:      game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
			wantOK:    true,
		},
		{
			name:      "excluded types",
			selection: compiler.StaticSelection{ExcludedTypes: []types.Card{types.Land}},
			want:      game.Selection{ExcludedTypes: []types.Card{types.Land}},
			wantOK:    true,
		},
		{
			name:      "supertypes",
			selection: compiler.StaticSelection{Supertypes: []types.Super{types.Legendary}},
			want:      game.Selection{Supertypes: []types.Super{types.Legendary}},
			wantOK:    true,
		},
		{
			name:      "excluded supertype single",
			selection: compiler.StaticSelection{ExcludedSupertypes: []types.Super{types.Legendary}},
			want:      game.Selection{ExcludedSupertype: types.Legendary},
			wantOK:    true,
		},
		{
			name:      "excluded supertype truncates to first",
			selection: compiler.StaticSelection{ExcludedSupertypes: []types.Super{types.Legendary, types.Basic}},
			want:      game.Selection{ExcludedSupertype: types.Legendary},
			wantOK:    true,
		},
		{
			name:      "subtypes any",
			selection: compiler.StaticSelection{SubtypesAny: []types.Sub{types.Goblin}},
			want:      game.Selection{SubtypesAny: []types.Sub{types.Goblin}},
			wantOK:    true,
		},
		{
			name:      "excluded subtype single",
			selection: compiler.StaticSelection{SubtypesAny: []types.Sub{types.Goblin}, ExcludedSubtypes: []types.Sub{types.Human}},
			want:      game.Selection{SubtypesAny: []types.Sub{types.Goblin}, ExcludedSubtype: types.Human},
			wantOK:    true,
		},
		{
			name:      "colors any",
			selection: compiler.StaticSelection{ColorsAny: []color.Color{color.Red}},
			want:      game.Selection{ColorsAny: []color.Color{color.Red}},
			wantOK:    true,
		},
		{
			name:      "colorless",
			selection: compiler.StaticSelection{Colorless: true},
			want:      game.Selection{Colorless: true},
			wantOK:    true,
		},
		{
			name:      "multicolored",
			selection: compiler.StaticSelection{Multicolored: true},
			want:      game.Selection{Multicolored: true},
			wantOK:    true,
		},
		{
			name:      "controller you",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, Controller: compiler.ControllerYou},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			wantOK:    true,
		},
		{
			name:      "controller opponent",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, Controller: compiler.ControllerOpponent},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent},
			wantOK:    true,
		},
		{
			name:      "controller not you",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, Controller: compiler.ControllerNotYou},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerNotYou},
			wantOK:    true,
		},
		{
			name:      "combat state attacking",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, CombatState: compiler.StaticCombatStateAttacking},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking},
			wantOK:    true,
		},
		{
			name:      "combat state blocking",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, CombatState: compiler.StaticCombatStateBlocking},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateBlocking},
			wantOK:    true,
		},
		{
			name:      "tap state tapped",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, TapState: compiler.StaticTapStateTapped},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue},
			wantOK:    true,
		},
		{
			name:      "tap state untapped",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, TapState: compiler.StaticTapStateUntapped},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriFalse},
			wantOK:    true,
		},
		{
			name:      "keyword",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, Keyword: parser.KeywordFlying},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying},
			wantOK:    true,
		},
		{
			name:      "excluded keyword",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: parser.KeywordFlying},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying},
			wantOK:    true,
		},
		{
			name:      "token only",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true},
			wantOK:    true,
		},
		{
			name:      "non token",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
			wantOK:    true,
		},

		// Per-context extras carried as a rider on the projector result.
		{
			name:      "match counter kind",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.Charge},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.Charge},
			wantOK:    true,
		},
		{
			name:      "match any counter",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, MatchAnyCounter: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchAnyCounter: true},
			wantOK:    true,
		},
		{
			name:      "subtype from entry choice",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, SubtypeFromEntryChoice: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry},
			wantOK:    true,
		},
		{
			name:      "color from entry choice",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, ColorFromEntryChoice: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorChoice: game.ColorChoiceSourceEntry},
			wantOK:    true,
		},
		{
			name:      "modified",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, Modified: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchModified: true},
			wantOK:    true,
		},
		{
			name:      "commander",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, Commander: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCommander: true},
			wantOK:    true,
		},
		{
			name:      "power threshold",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, MatchPower: true, Power: compare.Int{Op: compare.LessOrEqual, Value: 2}},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
			wantOK:    true,
		},
		{
			name:      "toughness threshold",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, MatchToughness: true, Toughness: compare.Int{Op: compare.GreaterOrEqual, Value: 3}},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Toughness: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})},
			wantOK:    true,
		},
		{
			name: "power or toughness disjunction",
			selection: compiler.StaticSelection{
				RequiredTypes:    []types.Card{types.Creature},
				PowerOrToughness: true,
				Power:            compare.Int{Op: compare.LessOrEqual, Value: 1},
				Toughness:        compare.Int{Op: compare.LessOrEqual, Value: 1},
			},
			want: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				AnyOf: []game.Selection{
					{Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 1})},
					{Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 1})},
				},
			},
			wantOK: true,
		},
		{
			name:      "power less than source",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, PowerLessThanSource: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, PowerLessThanSource: true},
			wantOK:    true,
		},
		{
			name:      "power greater than source",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, PowerGreaterThanSource: true},
			want:      game.Selection{RequiredTypes: []types.Card{types.Creature}, PowerGreaterThanSource: true},
			wantOK:    true,
		},

		// Fail-closed reject branches: an unknown clone-enum value must reject
		// rather than silently dropping the constraint.
		{
			name:      "unknown combat state rejects",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, CombatState: compiler.StaticCombatState(99)},
			wantOK:    false,
		},
		{
			name:      "unknown tap state rejects",
			selection: compiler.StaticSelection{RequiredTypes: []types.Card{types.Creature}, TapState: compiler.StaticTapState(99)},
			wantOK:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := lowerStaticSelection(tc.selection)
			if ok != tc.wantOK {
				t.Fatalf("lowerStaticSelection ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("lowerStaticSelection = %#v, want %#v", got, tc.want)
			}
		})
	}
}
