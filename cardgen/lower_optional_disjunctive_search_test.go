package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerDisjunctiveSearchFilterAnyOf proves a "creature or basic land card"
// library search lowers to a Selection.AnyOf of the two supertype-divergent
// disjuncts (a creature card and a basic land card) rather than a single
// flattened filter that would force every match to be both a creature/land and
// basic. The supertype "basic" attaches only to the land disjunct, so the two
// sides must stay separate alternatives.
func TestLowerDisjunctiveSearchFilterAnyOf(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Disjunctive Tutor",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Sorcery",
		OracleText: "Search your library for a creature or basic land card, reveal it, put it into your hand, then shuffle.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	search, ok := mode.Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive = %#v, want Search", mode.Sequence[0].Primitive)
	}
	alts := search.Spec.Filter.AnyOf
	if len(alts) != 2 {
		t.Fatalf("filter alternatives = %#v, want two", search.Spec.Filter)
	}
	creature := alts[0]
	if len(creature.RequiredTypes) != 1 || creature.RequiredTypes[0] != types.Creature ||
		len(creature.Supertypes) != 0 {
		t.Errorf("first alternative = %#v, want a plain creature filter", creature)
	}
	land := alts[1]
	if len(land.RequiredTypes) != 1 || land.RequiredTypes[0] != types.Land ||
		len(land.Supertypes) != 1 || land.Supertypes[0] != types.Basic {
		t.Errorf("second alternative = %#v, want a basic land filter", land)
	}
}

// TestLowerSubtypeUnionSearchFilterStaysFlat proves a "basic Forest or Plains
// card" search keeps its flattened single filter (basic supertype plus a
// Forest-or-Plains subtype union) and is NOT split into alternatives. Here the
// leading "basic" distributes across both subtypes, so the flattened parse is
// correct and splitting would wrongly drop "basic" from the Plains side.
func TestLowerSubtypeUnionSearchFilterStaysFlat(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Dual Land Tutor",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Sorcery",
		OracleText: "Search your library for a basic Forest or Plains card, put it onto the battlefield tapped, then shuffle.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	search, ok := mode.Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive = %#v, want Search", mode.Sequence[0].Primitive)
	}
	if len(search.Spec.Filter.AnyOf) != 0 {
		t.Fatalf("filter should stay flat, got AnyOf %#v", search.Spec.Filter.AnyOf)
	}
	filter := search.Spec.Filter
	if len(filter.Supertypes) != 1 || filter.Supertypes[0] != types.Basic {
		t.Errorf("filter supertypes = %#v, want basic", filter.Supertypes)
	}
	if len(filter.SubtypesAny) != 2 {
		t.Errorf("filter subtypes = %#v, want Forest or Plains", filter.SubtypesAny)
	}
}

// TestLowerOptionalResolvingChapterDisjunctiveSearch proves a Saga chapter that
// pairs an optional resolving effect with a disjunctive search lowers fully: the
// optional sacrifice publishes its result and the following search is gated on
// that result having succeeded (The Huntsman's Redemption chapter II).
func TestLowerOptionalResolvingChapterDisjunctiveSearch(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Redemption Saga",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "I — Create a 3/3 green Beast creature token.\n" +
			"II — You may sacrifice a creature. If you do, search your library for a creature or basic land card, reveal it, put it into your hand, then shuffle.\n" +
			"III — Draw a card.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
	sequence := face.ChapterAbilities[1].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("chapter II sequence length = %d, want 2", len(sequence))
	}
	if _, ok := sequence[0].Primitive.(game.SacrificePermanents); !ok {
		t.Fatalf("first primitive = %#v, want SacrificePermanents", sequence[0].Primitive)
	}
	if !sequence[0].Optional || sequence[0].PublishResult == "" {
		t.Errorf("first instruction should be optional and publish a result, got %#v", sequence[0])
	}
	search, ok := sequence[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("second primitive = %#v, want Search", sequence[1].Primitive)
	}
	if len(search.Spec.Filter.AnyOf) != 2 {
		t.Errorf("search filter = %#v, want two alternatives", search.Spec.Filter)
	}
	if !sequence[1].ResultGate.Exists ||
		sequence[1].ResultGate.Val.Key != sequence[0].PublishResult ||
		sequence[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Errorf("search should be gated on the optional sacrifice succeeding, got %#v", sequence[1].ResultGate)
	}
}
