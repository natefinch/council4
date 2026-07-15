package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// finaleOracleText is Finale of Devastation's remaining-mechanics oracle wording,
// the exact composite this recognizer lowers.
const finaleOracleText = "Search your library and/or graveyard for a creature card with mana value X or less and put it onto the battlefield. If you search your library this way, shuffle. If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn."

// lowerFinaleSequence lowers a Finale-shaped sorcery and returns its three
// lowered instructions, asserting the single-mode/no-target sequence shape the
// multi-zone search recognizer must produce.
func lowerFinaleSequence(t *testing.T, oracle string) []game.Instruction {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Finale",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}{G}",
		OracleText: oracle,
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability = none, want a lowered spell")
	}
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(content.Modes))
	}
	if len(content.Modes[0].Targets) != 0 {
		t.Fatalf("targets = %#v, want none", content.Modes[0].Targets)
	}
	seq := content.Modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence = %d instructions, want 3", len(seq))
	}
	return seq
}

// TestLowerFinaleOfDevastationMultiZoneSearch proves the composite Finale of
// Devastation wording lowers to the three-instruction shape: a folded multi-zone
// battlefield search that records whether the library was searched, a
// ShuffleLibrary gated on that recorded result, and an X>=10 group +X/+X and
// haste rider. Every consumed field is typed, so the runtime never inspects card
// text or name.
func TestLowerFinaleOfDevastationMultiZoneSearch(t *testing.T) {
	t.Parallel()
	seq := lowerFinaleSequence(t, finaleOracleText)

	// Instruction 0: the folded multi-zone search publishes the searched-library
	// result and carries the creature/mana-value<=X filter to the battlefield.
	search, ok := seq[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("instruction 0 = %T, want game.Search", seq[0].Primitive)
	}
	if seq[0].PublishResult != searchedLibraryResultKey {
		t.Fatalf("search publishes %q, want %q", seq[0].PublishResult, searchedLibraryResultKey)
	}
	if !search.Spec.AlsoGraveyard {
		t.Fatal("AlsoGraveyard = false, want true")
	}
	if !search.Spec.ConditionalShuffle {
		t.Fatal("ConditionalShuffle = false, want true")
	}
	if search.Spec.Destination != zone.Battlefield {
		t.Fatalf("destination = %v, want battlefield", search.Spec.Destination)
	}
	if search.Spec.SourceZone != zone.Library {
		t.Fatalf("source zone = %v, want library", search.Spec.SourceZone)
	}
	if !search.Spec.MaxManaValueFromX {
		t.Fatal("MaxManaValueFromX = false, want true (mana value X or less)")
	}
	if !slices.Equal(search.Spec.Filter.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("filter required types = %#v, want creature", search.Spec.Filter.RequiredTypes)
	}
	if got := search.Amount; got != game.Fixed(1) {
		t.Fatalf("amount = %#v, want fixed 1", got)
	}

	// Instruction 1: the shuffle is gated on the searched-library result, so it
	// runs exactly when the library was among the searched zones and never for a
	// graveyard-only search.
	shuffle, ok := seq[1].Primitive.(game.ShuffleLibrary)
	if !ok {
		t.Fatalf("instruction 1 = %T, want game.ShuffleLibrary", seq[1].Primitive)
	}
	_ = shuffle
	if !seq[1].ResultGate.Exists {
		t.Fatal("shuffle result gate = none, want searched-library gate")
	}
	gate := seq[1].ResultGate.Val
	if gate.Key != searchedLibraryResultKey {
		t.Fatalf("gate key = %q, want %q", gate.Key, searchedLibraryResultKey)
	}
	if gate.SearchedLibrary != game.TriTrue {
		t.Fatalf("gate SearchedLibrary = %v, want TriTrue", gate.SearchedLibrary)
	}
	if gate.Succeeded != game.TriAny || gate.Accepted != game.TriAny {
		t.Fatalf("gate = %#v, want only SearchedLibrary constrained", gate)
	}

	// Instruction 2: the X>=10 rider grants dynamic +X/+X and haste until end of
	// turn to the resolving controller's creatures.
	apply, ok := seq[2].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("instruction 2 = %T, want game.ApplyContinuous", seq[2].Primitive)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply duration = %v, want until end of turn", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 2 {
		t.Fatalf("continuous effects = %d, want 2 (pump + haste)", len(apply.ContinuousEffects))
	}
	pump := apply.ContinuousEffects[0]
	if pump.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("pump layer = %v, want power/toughness modify", pump.Layer)
	}
	for _, side := range []struct {
		name string
		dyn  game.DynamicAmount
		ok   bool
	}{
		{"power", pump.PowerDeltaDynamic.Val, pump.PowerDeltaDynamic.Exists},
		{"toughness", pump.ToughnessDeltaDynamic.Val, pump.ToughnessDeltaDynamic.Exists},
	} {
		if !side.ok {
			t.Fatalf("%s delta dynamic = none, want dynamic +X", side.name)
		}
		if side.dyn.Kind != game.DynamicAmountX {
			t.Fatalf("%s delta kind = %v, want DynamicAmountX", side.name, side.dyn.Kind)
		}
		if side.dyn.Multiplier != 1 {
			t.Fatalf("%s delta multiplier = %d, want +1 (positive +X/+X)", side.name, side.dyn.Multiplier)
		}
	}
	haste := apply.ContinuousEffects[1]
	if haste.Layer != game.LayerAbility ||
		len(haste.AddKeywords) != 1 ||
		haste.AddKeywords[0] != game.Haste {
		t.Fatalf("haste layer = %#v, want add haste", haste)
	}

	// The rider gate is the X>=10 effect condition read from the resolving spell's
	// chosen X.
	if !seq[2].Condition.Exists || !seq[2].Condition.Val.Condition.Exists {
		t.Fatal("apply condition = none, want X>=10 gate")
	}
	aggs := seq[2].Condition.Val.Condition.Val.Aggregates
	if len(aggs) != 1 {
		t.Fatalf("aggregates = %#v, want one X>=10 comparison", aggs)
	}
	if aggs[0].Aggregate != game.AggregateSpellX ||
		aggs[0].Op != compare.GreaterOrEqual ||
		aggs[0].Value != 10 {
		t.Fatalf("aggregate = %#v, want spell X >= 10", aggs[0])
	}
}

// TestRenderFinaleOfDevastationRoundTrip proves the full compile -> lower ->
// render pipeline emits valid Go source carrying the new typed fields, so the
// generated card round-trips. Rendering is explicit field-by-field, so a missing
// render line for ConditionalShuffle or SearchedLibrary would silently drop the
// field; asserting their presence guards that.
func TestRenderFinaleOfDevastationRoundTrip(t *testing.T) {
	t.Parallel()
	source := generateExecutable(t, &ScryfallCard{
		Name:       "Finale of Devastation",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}{G}",
		Colors:     []string{"G"},
		OracleText: finaleOracleText,
	})
	for _, want := range []string{
		"game.Search{",
		"AlsoGraveyard:",
		"ConditionalShuffle: true,",
		"MaxManaValueFromX:",
		"game.ShuffleLibrary{",
		"SearchedLibrary: game.TriTrue,",
		"game.ApplyContinuous{",
		"game.Haste",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("rendered source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerMultiZoneSearchWrongDestinationFailsClosed proves a multi-zone search
// whose found card is put somewhere other than the battlefield is not lowered by
// this recognizer, keeping unsupported destinations out of the corpus.
func TestLowerMultiZoneSearchWrongDestinationFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Finale Hand",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}{G}",
		OracleText: "Search your library and/or graveyard for a creature card with mana value X or less and put it into your hand. If you search your library this way, shuffle. If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn.",
	})
}

// TestLowerMultiZoneSearchUnconditionalShuffleFailsClosed proves that dropping the
// "If you search your library this way" gate on the shuffle (making it
// unconditional) fails closed rather than shuffling on a graveyard-only search.
func TestLowerMultiZoneSearchUnconditionalShuffleFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Finale No Gate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}{G}",
		OracleText: "Search your library and/or graveyard for a creature card with mana value X or less and put it onto the battlefield, then shuffle. If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn.",
	})
}

// TestLowerMultiZoneSearchFixedManaValueFailsClosed proves the dynamic {X}
// mana-value bound is required: a fixed "mana value 3 or less" multi-zone search
// is not lowered by this recognizer.
func TestLowerMultiZoneSearchFixedManaValueFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Finale Fixed MV",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}{G}",
		OracleText: "Search your library and/or graveyard for a creature card with mana value 3 or less and put it onto the battlefield. If you search your library this way, shuffle. If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn.",
	})
}
