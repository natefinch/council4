package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// revealChooseDiscardPrimitive lowers a "Target opponent reveals their hand. You
// choose a [filter] card from it. That player discards that card." spell and
// returns its single ChooseDiscardFromHand instruction so the typed filter can
// be asserted.
func revealChooseDiscardPrimitive(t *testing.T, name, oracle string) game.ChooseDiscardFromHand {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       name,
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Sorcery",
		OracleText: oracle,
		Colors:     []string{"B"},
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want a single instruction", mode.Sequence)
	}
	choose, ok := mode.Sequence[0].Primitive.(game.ChooseDiscardFromHand)
	if !ok {
		t.Fatalf("instruction[0] = %T, want game.ChooseDiscardFromHand", mode.Sequence[0].Primitive)
	}
	if choose.Player != game.TargetPlayerReference(0) {
		t.Fatalf("Player = %#v, want target player 0", choose.Player)
	}
	return choose
}

// TestLowerRevealChooseDiscardPositiveCardType proves a single positive card-type
// descriptor ("a creature card", "an artifact card") lowers to a
// ChooseDiscardFromHand whose Selection requires that type (Ostracize, Shattered
// Dreams). The exclude flags stay clear because the filter is purely positive.
func TestLowerRevealChooseDiscardPositiveCardType(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		oracle string
		want   types.Card
	}{
		{"Ostracize", "Target opponent reveals their hand. You choose a creature card from it. That player discards that card.", types.Creature},
		{"Shattered Dreams", "Target opponent reveals their hand. You choose an artifact card from it. That player discards that card.", types.Artifact},
	} {
		choose := revealChooseDiscardPrimitive(t, tc.name, tc.oracle)
		if choose.ExcludeCreature || choose.ExcludeLand || choose.MaxManaValue.Exists {
			t.Fatalf("%s: exclude flags set: %#v", tc.name, choose)
		}
		if !slices.Equal(choose.Selection.RequiredTypesAny, []types.Card{tc.want}) {
			t.Fatalf("%s: RequiredTypesAny = %#v, want [%s]", tc.name, choose.Selection.RequiredTypesAny, tc.want)
		}
		if choose.Selection.ExcludedSupertype != "" {
			t.Fatalf("%s: ExcludedSupertype = %q, want empty", tc.name, choose.Selection.ExcludedSupertype)
		}
	}
}

// TestLowerRevealChooseDiscardTypeUnion proves an "or"-joined card-type union
// ("a creature or planeswalker card", "an artifact or creature card") lowers to a
// disjunctive RequiredTypesAny set (Despise, Divest).
func TestLowerRevealChooseDiscardTypeUnion(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		oracle string
		want   []types.Card
	}{
		{"Despise", "Target opponent reveals their hand. You choose a creature or planeswalker card from it. That player discards that card.", []types.Card{types.Creature, types.Planeswalker}},
		{"Divest", "Target player reveals their hand. You choose an artifact or creature card from it. That player discards that card.", []types.Card{types.Artifact, types.Creature}},
	} {
		choose := revealChooseDiscardPrimitive(t, tc.name, tc.oracle)
		if !slices.Equal(choose.Selection.RequiredTypesAny, tc.want) {
			t.Fatalf("%s: RequiredTypesAny = %#v, want %#v", tc.name, choose.Selection.RequiredTypesAny, tc.want)
		}
	}
}

// TestLowerRevealChooseDiscardSupertypeExclusion proves the "nonbasic" and
// "nonlegendary" supertype exclusions lower to ExcludedSupertype, composing with
// a required type (Encroach's "nonbasic land") or with the nonland exclude flag
// (Lay Bare the Heart's "nonlegendary, nonland").
func TestLowerRevealChooseDiscardSupertypeExclusion(t *testing.T) {
	t.Parallel()

	encroach := revealChooseDiscardPrimitive(t, "Encroach",
		"Target player reveals their hand. You choose a nonbasic land card from it. That player discards that card.")
	if !slices.Equal(encroach.Selection.RequiredTypesAny, []types.Card{types.Land}) {
		t.Fatalf("Encroach: RequiredTypesAny = %#v, want [Land]", encroach.Selection.RequiredTypesAny)
	}
	if encroach.Selection.ExcludedSupertype != types.Basic {
		t.Fatalf("Encroach: ExcludedSupertype = %q, want Basic", encroach.Selection.ExcludedSupertype)
	}

	layBare := revealChooseDiscardPrimitive(t, "Lay Bare the Heart",
		"Target opponent reveals their hand. You choose a nonlegendary, nonland card from it. That player discards that card.")
	if !layBare.ExcludeLand {
		t.Fatal("Lay Bare the Heart: ExcludeLand = false, want true")
	}
	if len(layBare.Selection.RequiredTypesAny) != 0 {
		t.Fatalf("Lay Bare the Heart: RequiredTypesAny = %#v, want empty", layBare.Selection.RequiredTypesAny)
	}
	if layBare.Selection.ExcludedSupertype != types.Legendary {
		t.Fatalf("Lay Bare the Heart: ExcludedSupertype = %q, want Legendary", layBare.Selection.ExcludedSupertype)
	}
}

// TestLowerRevealChooseDiscardExcludeOnlyHasNoSelection proves the established
// "noncreature, nonland" descriptor (Duress) still lowers to bare exclude flags
// with no Selection, so the broadening leaves the existing supported output
// byte-identical.
func TestLowerRevealChooseDiscardExcludeOnlyHasNoSelection(t *testing.T) {
	t.Parallel()
	choose := revealChooseDiscardPrimitive(t, "Duress",
		"Target opponent reveals their hand. You choose a noncreature, nonland card from it. That player discards that card.")
	if !choose.ExcludeCreature || !choose.ExcludeLand {
		t.Fatalf("Duress: exclude flags = %#v, want both set", choose)
	}
	if len(choose.Selection.RequiredTypesAny) != 0 || choose.Selection.ExcludedSupertype != "" {
		t.Fatalf("Duress: Selection = %#v, want zero", choose.Selection)
	}
}

// TestLowerRevealChooseDiscardConjunctiveTypeLineFailsClosed proves an adjacent
// (non-"or") two-type descriptor fails closed rather than being mistaken for a
// disjunctive union: "artifact creature" is a conjunctive type line the filter
// cannot express, so the whole card stays unsupported.
func TestLowerRevealChooseDiscardConjunctiveTypeLineFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Conjunctive Probe",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Sorcery",
		OracleText: "Target opponent reveals their hand. You choose an artifact creature card from it. That player discards that card.",
		Colors:     []string{"B"},
	})
}
