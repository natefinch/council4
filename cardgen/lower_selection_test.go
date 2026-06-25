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
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// permanentSelector returns the minimal selector that projects to an empty
// Selection: a bare "permanent" with no qualifiers. Dimension tests start from
// it and switch on a single dimension so the expected Selection isolates that
// dimension's mapping.
func permanentSelector() compiler.CompiledSelector {
	return compiler.CompiledSelector{Kind: compiler.SelectorPermanent}
}

// TestSelectionForSelectorDimensions maps every CompiledSelector dimension the
// canonical projector honors onto its game.Selection field. Each case sets one
// dimension on a bare permanent selector and asserts the whole projected
// Selection, so a mapping that writes the wrong field (or an extra one) fails.
func TestSelectionForSelectorDimensions(t *testing.T) {
	cases := []struct {
		name string
		sel  compiler.CompiledSelector
		want game.Selection
	}{
		{
			name: "required name",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorCreature, RequiredName: "Charmed Stray"},
			want: game.Selection{RequiredTypes: []types.Card{types.Creature}, Name: "Charmed Stray"},
		},
		{
			name: "bare permanent",
			sel:  permanentSelector(),
			want: game.Selection{},
		},
		{
			name: "kind creature",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorCreature},
			want: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		},
		{
			name: "kind artifact",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorArtifact},
			want: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
		},
		{
			name: "kind enchantment",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorEnchantment},
			want: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
		},
		{
			name: "kind land",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorLand},
			want: game.Selection{RequiredTypes: []types.Card{types.Land}},
		},
		{
			name: "kind planeswalker",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPlaneswalker},
			want: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
		},
		{
			name: "kind battle",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorBattle},
			want: game.Selection{RequiredTypes: []types.Card{types.Battle}},
		},
		{
			name: "kind unknown with subtype",
			sel: compiler.CompiledSelector{Kind: compiler.SelectorUnknown}.
				WithAtoms(compiler.CompiledSelectorAtoms{SubtypesAny: []types.Sub{types.Island}}),
			want: game.Selection{SubtypesAny: []types.Sub{types.Island}},
		},
		{
			name: "required types any",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{RequiredTypesAny: []types.Card{types.Creature, types.Artifact}}),
			want: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Artifact}},
		},
		{
			name: "excluded types",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedTypes: []types.Card{types.Artifact}}),
			want: game.Selection{ExcludedTypes: []types.Card{types.Artifact}},
		},
		{
			name: "supertypes",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{Supertypes: []types.Super{types.Legendary}}),
			want: game.Selection{Supertypes: []types.Super{types.Legendary}},
		},
		{
			name: "subtypes any",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{SubtypesAny: []types.Sub{types.Goblin}}),
			want: game.Selection{SubtypesAny: []types.Sub{types.Goblin}},
		},
		{
			name: "colors any",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ColorsAny: []color.Color{color.Red}}),
			want: game.Selection{ColorsAny: []color.Color{color.Red}},
		},
		{
			name: "excluded colors",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedColors: []color.Color{color.White}}),
			want: game.Selection{ExcludedColors: []color.Color{color.White}},
		},
		{
			name: "colorless",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Colorless: true},
			want: game.Selection{Colorless: true},
		},
		{
			name: "multicolored",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Multicolored: true},
			want: game.Selection{Multicolored: true},
		},
		{
			name: "entered this turn",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, EnteredThisTurn: true},
			want: game.Selection{EnteredThisTurn: true},
		},
		{
			name: "controller you",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Controller: compiler.ControllerYou},
			want: game.Selection{Controller: game.ControllerYou},
		},
		{
			name: "controller opponent",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Controller: compiler.ControllerOpponent},
			want: game.Selection{Controller: game.ControllerOpponent},
		},
		{
			name: "controller not you",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Controller: compiler.ControllerNotYou},
			want: game.Selection{Controller: game.ControllerNotYou},
		},
		{
			name: "attacking",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Attacking: true},
			want: game.Selection{CombatState: game.CombatStateAttacking},
		},
		{
			name: "blocking",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Blocking: true},
			want: game.Selection{CombatState: game.CombatStateBlocking},
		},
		{
			name: "attacking or blocking",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Attacking: true, Blocking: true},
			want: game.Selection{CombatState: game.CombatStateAttackingOrBlocking},
		},
		{
			name: "tapped",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Tapped: true},
			want: game.Selection{Tapped: game.TriTrue},
		},
		{
			name: "untapped",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Untapped: true},
			want: game.Selection{Tapped: game.TriFalse},
		},
		{
			name: "mana value",
			sel: compiler.CompiledSelector{
				Kind:           compiler.SelectorPermanent,
				MatchManaValue: true,
				ManaValue:      compare.Int{Op: compare.LessOrEqual, Value: 3},
			},
			want: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})},
		},
		{
			name: "power",
			sel: compiler.CompiledSelector{
				Kind:       compiler.SelectorPermanent,
				MatchPower: true,
				Power:      compare.Int{Op: compare.GreaterOrEqual, Value: 2},
			},
			want: game.Selection{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2})},
		},
		{
			name: "toughness",
			sel: compiler.CompiledSelector{
				Kind:           compiler.SelectorPermanent,
				MatchToughness: true,
				Toughness:      compare.Int{Op: compare.LessThan, Value: 4},
			},
			want: game.Selection{Toughness: opt.Val(compare.Int{Op: compare.LessThan, Value: 4})},
		},
		{
			name: "keyword",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Keyword: parser.KeywordFlying},
			want: game.Selection{Keyword: game.Flying},
		},
		{
			name: "excluded keyword",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, ExcludedKeyword: parser.KeywordFlying},
			want: game.Selection{ExcludedKeyword: game.Flying},
		},
		{
			name: "match counter",
			sel: compiler.CompiledSelector{
				Kind:            compiler.SelectorPermanent,
				MatchCounter:    true,
				RequiredCounter: counter.PlusOnePlusOne,
			},
			want: game.Selection{MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne},
		},
		{
			name: "subtype from entry choice",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, SubtypeFromEntryChoice: true},
			want: game.Selection{SubtypeChoice: game.SubtypeChoiceSourceEntry},
		},
		{
			name: "subtype from chosen type",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, SubtypeFromChosenType: true},
			want: game.Selection{SubtypeChoice: game.SubtypeChoiceResolution},
		},
		{
			name: "exclude source via another",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Another: true},
			want: game.Selection{ExcludeSource: true},
		},
		{
			name: "exclude source via other",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Other: true},
			want: game.Selection{ExcludeSource: true},
		},
		{
			name: "excluded supertype",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedSupertypes: []types.Super{types.Legendary}}),
			want: game.Selection{ExcludedSupertype: types.Legendary},
		},
		{
			name: "excluded subtype",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedSubtypes: []types.Sub{types.Human}}),
			want: game.Selection{ExcludedSubtype: types.Human},
		},
		{
			name: "match any counter",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, MatchAnyCounter: true},
			want: game.Selection{MatchAnyCounter: true},
		},
		{
			name: "subtype from chosen type excluded",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, SubtypeFromChosenTypeExcluded: true},
			want: game.Selection{SubtypeChoice: game.SubtypeChoiceResolutionExcluded},
		},
		{
			name: "conjunctive types",
			sel: compiler.CompiledSelector{Kind: compiler.SelectorPermanent, ConjunctiveTypes: true}.
				WithAtoms(compiler.CompiledSelectorAtoms{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
			want: game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
		},
		{
			name: "non token",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, NonToken: true},
			want: game.Selection{NonToken: true},
		},
		{
			name: "token only",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, TokenOnly: true},
			want: game.Selection{TokenOnly: true},
		},
		{
			name: "historic",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Historic: true},
			want: game.Selection{AnyOf: []game.Selection{
				{RequiredTypes: []types.Card{types.Artifact}},
				{Supertypes: []types.Super{types.Legendary}},
				{SubtypesAny: []types.Sub{types.Saga}},
			}},
		},
		{
			name: "power less than source",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, PowerLessThanSource: true},
			want: game.Selection{PowerLessThanSource: true},
		},
		{
			name: "power greater than source",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, PowerGreaterThanSource: true},
			want: game.Selection{PowerGreaterThanSource: true},
		},
		{
			name: "name unique among controlled",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, NameUniqueAmongControlled: true},
			want: game.Selection{NameUniqueAmongControlled: true},
		},
		{
			name: "match no counters",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, MatchNoCounters: true},
			want: game.Selection{MatchNoCounters: true},
		},
		{
			name: "match excluded counter",
			sel: compiler.CompiledSelector{
				Kind:                 compiler.SelectorPermanent,
				MatchExcludedCounter: true,
				ExcludedCounter:      counter.Charge,
			},
			want: game.Selection{MatchExcludedCounter: true, ExcludedCounter: counter.Charge},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := SelectionForSelector(tc.sel)
			if !ok {
				t.Fatal("SelectionForSelector ok = false, want true")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Selection mismatch\n got: %#v\nwant: %#v", got, tc.want)
			}
		})
	}
}

// TestSelectionForSelectorHardRejects covers every predicate dimension that no
// game.Selection field can carry. The canonical projector must fail closed on
// each so an unsupported wording is never silently dropped.
func TestSelectionForSelectorHardRejects(t *testing.T) {
	cases := []struct {
		name string
		sel  compiler.CompiledSelector
	}{
		{
			name: "zone",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Zone: zone.Graveyard},
		},
		{
			name: "basic land type",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, BasicLandType: true},
		},
		{
			name: "player or planeswalker",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, PlayerOrPlaneswalker: true},
		},
		{
			name: "match total mana value",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, MatchTotalManaValue: true},
		},
		{
			name: "inclusive one of each",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, InclusiveOneOfEach: true},
		},
		{
			name: "source types",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{SourceTypes: []types.Card{types.Artifact}}),
		},
		{
			name: "tapped and untapped",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Tapped: true, Untapped: true},
		},
		{
			name: "mana value from x",
			sel: compiler.CompiledSelector{
				Kind:           compiler.SelectorPermanent,
				MatchManaValue: true,
				ManaValueX:     true,
				ManaValue:      compare.Int{Op: compare.LessOrEqual},
			},
		},
		{
			name: "unknown kind without subtype",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorUnknown},
		},
		{
			name: "invalid controller",
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Controller: compiler.ControllerKind(99)},
		},
		{
			name: "two excluded supertypes",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedSupertypes: []types.Super{types.Legendary, types.Basic}}),
		},
		{
			name: "two excluded subtypes",
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedSubtypes: []types.Sub{types.Human, types.Goblin}}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := SelectionForSelector(tc.sel)
			if ok {
				t.Fatalf("SelectionForSelector ok = true, want false (got %#v)", got)
			}
			if !reflect.DeepEqual(got, game.Selection{}) {
				t.Fatalf("rejected projection must be zero Selection, got %#v", got)
			}
		})
	}
}

// maskableDim pairs a maskable dimension with a selector that activates it and
// the Selection field that dimension produces when honored.
type maskableDim struct {
	name string
	dim  SelectionDim
	sel  compiler.CompiledSelector
	want game.Selection
}

func maskableDims() []maskableDim {
	return []maskableDim{
		{
			name: "exclude source",
			dim:  DimExcludeSource,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Another: true},
			want: game.Selection{ExcludeSource: true},
		},
		{
			name: "excluded supertype",
			dim:  DimExcludedSupertype,
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedSupertypes: []types.Super{types.Legendary}}),
			want: game.Selection{ExcludedSupertype: types.Legendary},
		},
		{
			name: "excluded subtype",
			dim:  DimExcludedSubtype,
			sel: permanentSelector().
				WithAtoms(compiler.CompiledSelectorAtoms{ExcludedSubtypes: []types.Sub{types.Human}}),
			want: game.Selection{ExcludedSubtype: types.Human},
		},
		{
			name: "match any counter",
			dim:  DimMatchAnyCounter,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, MatchAnyCounter: true},
			want: game.Selection{MatchAnyCounter: true},
		},
		{
			name: "subtype choice excluded",
			dim:  DimSubtypeChoiceExcluded,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, SubtypeFromChosenTypeExcluded: true},
			want: game.Selection{SubtypeChoice: game.SubtypeChoiceResolutionExcluded},
		},
		{
			name: "conjunctive types",
			dim:  DimConjunctiveTypes,
			sel: compiler.CompiledSelector{Kind: compiler.SelectorPermanent, ConjunctiveTypes: true}.
				WithAtoms(compiler.CompiledSelectorAtoms{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
			want: game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
		},
		{
			name: "non token",
			dim:  DimNonToken,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, NonToken: true},
			want: game.Selection{NonToken: true},
		},
		{
			name: "token only",
			dim:  DimTokenOnly,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, TokenOnly: true},
			want: game.Selection{TokenOnly: true},
		},
		{
			name: "historic",
			dim:  DimHistoric,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, Historic: true},
			want: game.Selection{AnyOf: []game.Selection{
				{RequiredTypes: []types.Card{types.Artifact}},
				{Supertypes: []types.Super{types.Legendary}},
				{SubtypesAny: []types.Sub{types.Saga}},
			}},
		},
		{
			name: "power vs source",
			dim:  DimPowerVsSource,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, PowerLessThanSource: true},
			want: game.Selection{PowerLessThanSource: true},
		},
		{
			name: "required name",
			dim:  DimRequiredName,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, RequiredName: "Charmed Stray"},
			want: game.Selection{Name: "Charmed Stray"},
		},
		{
			name: "name unique among controlled",
			dim:  DimNameUniqueAmongControlled,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, NameUniqueAmongControlled: true},
			want: game.Selection{NameUniqueAmongControlled: true},
		},
		{
			name: "match no counters",
			dim:  DimMatchNoCounters,
			sel:  compiler.CompiledSelector{Kind: compiler.SelectorPermanent, MatchNoCounters: true},
			want: game.Selection{MatchNoCounters: true},
		},
		{
			name: "match excluded counter",
			dim:  DimMatchExcludedCounter,
			sel: compiler.CompiledSelector{
				Kind:                 compiler.SelectorPermanent,
				MatchExcludedCounter: true,
				ExcludedCounter:      counter.Charge,
			},
			want: game.Selection{MatchExcludedCounter: true, ExcludedCounter: counter.Charge},
		},
	}
}

// TestSelectionMaskHonor confirms an empty mask honors every maskable dimension,
// reproducing the canonical superset.
func TestSelectionMaskHonor(t *testing.T) {
	for _, d := range maskableDims() {
		t.Run(d.name, func(t *testing.T) {
			got, ok := SelectionForSelectorMasked(d.sel, SelectionMask{})
			if !ok {
				t.Fatal("honored dimension ok = false, want true")
			}
			if !reflect.DeepEqual(got, d.want) {
				t.Fatalf("honored projection mismatch\n got: %#v\nwant: %#v", got, d.want)
			}
		})
	}
}

// TestSelectionMaskReject confirms a Rejecting mask fails the projection closed
// when the rejected dimension is active, reproducing a projector that historically
// failed closed on that qualifier.
func TestSelectionMaskReject(t *testing.T) {
	for _, d := range maskableDims() {
		t.Run(d.name, func(t *testing.T) {
			got, ok := SelectionForSelectorMasked(d.sel, SelectionMask{}.Rejecting(d.dim))
			if ok {
				t.Fatalf("rejected dimension ok = true, want false (got %#v)", got)
			}
			if !reflect.DeepEqual(got, game.Selection{}) {
				t.Fatalf("rejected projection must be zero Selection, got %#v", got)
			}
		})
	}
}

// TestSelectionMaskIgnore confirms an Ignoring mask drops the dimension's field
// while keeping the projection valid, reproducing a projector that historically
// dropped a qualifier its context never carried.
func TestSelectionMaskIgnore(t *testing.T) {
	for _, d := range maskableDims() {
		t.Run(d.name, func(t *testing.T) {
			got, ok := SelectionForSelectorMasked(d.sel, SelectionMask{}.Ignoring(d.dim))
			if !ok {
				t.Fatal("ignored dimension ok = false, want true")
			}
			want := game.Selection{}
			if d.dim == DimConjunctiveTypes {
				// Ignoring the conjunctive routing leaves the type set on the
				// default any-of union rather than dropping it.
				want = game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ignored projection mismatch\n got: %#v\nwant: %#v", got, want)
			}
		})
	}
}

// TestSelectionMaskInactiveDimension confirms masking a dimension that is not
// active in the selector changes nothing: an inactive rejected dimension still
// projects successfully.
func TestSelectionMaskInactiveDimension(t *testing.T) {
	sel := compiler.CompiledSelector{Kind: compiler.SelectorCreature}
	mask := SelectionMask{}.Rejecting(
		DimExcludeSource,
		DimExcludedSupertype,
		DimExcludedSubtype,
		DimMatchAnyCounter,
		DimSubtypeChoiceExcluded,
		DimConjunctiveTypes,
		DimNonToken,
		DimTokenOnly,
		DimHistoric,
		DimPowerVsSource,
	)
	got, ok := SelectionForSelectorMasked(sel, mask)
	if !ok {
		t.Fatal("inactive rejected dimensions ok = false, want true")
	}
	want := game.Selection{RequiredTypes: []types.Card{types.Creature}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("projection mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
