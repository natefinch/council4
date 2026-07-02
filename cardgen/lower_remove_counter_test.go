package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// removeCounterFromMode pulls the single RemoveCounter primitive out of a
// one-target, one-instruction ability mode, failing the test on any other
// shape.
func removeCounterFromMode(t *testing.T, content game.AbilityContent) game.RemoveCounter {
	t.Helper()
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %#v", content.Modes)
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowPermanent {
		t.Fatalf("targets = %#v", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	remove, ok := mode.Sequence[0].Primitive.(game.RemoveCounter)
	if !ok {
		t.Fatalf("primitive = %#v", mode.Sequence[0].Primitive)
	}
	if remove.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %#v", remove.Object)
	}
	return remove
}

// TestLowerRemoveCounterTriggeredChooseKind proves the kind-unspecified
// "remove a counter from target permanent" trigger body (Ferropede) lowers to a
// RemoveCounter that leaves the kind to the resolving controller (ChooseKind).
func TestLowerRemoveCounterTriggeredChooseKind(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ferropede",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Insect",
		OracleText: "Whenever this creature deals combat damage to a player, you may remove a counter from target permanent.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v", face.TriggeredAbilities)
	}
	remove := removeCounterFromMode(t, face.TriggeredAbilities[0].Content)
	if remove.Amount != game.Fixed(1) || !remove.ChooseKind || remove.CounterKind != 0 {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveCounterEnterChooseKind proves the enters-trigger form
// (Medicine Runner) lowers the same kind-unspecified removal.
func TestLowerRemoveCounterEnterChooseKind(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Medicine Runner",
		Layout:     "normal",
		TypeLine:   "Creature — Rabbit Scout",
		OracleText: "When this creature enters, you may remove a counter from target permanent.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v", face.TriggeredAbilities)
	}
	remove := removeCounterFromMode(t, face.TriggeredAbilities[0].Content)
	if remove.Amount != game.Fixed(1) || !remove.ChooseKind {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveCounterActivatedNamedKind proves the activated form with a
// named kind ("{3}, {T}: Remove a -1/-1 counter from target creature.",
// Chainbreaker) lowers to a RemoveCounter naming that exact kind, not a choice.
func TestLowerRemoveCounterActivatedNamedKind(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Chainbreaker",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Scarecrow",
		OracleText: "{3}, {T}: Remove a -1/-1 counter from target creature.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v", face.ActivatedAbilities)
	}
	remove := removeCounterFromMode(t, face.ActivatedAbilities[0].Content)
	if remove.Amount != game.Fixed(1) || remove.ChooseKind || remove.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveCounterNonlandTarget proves the "target nonland permanent"
// restriction (Thrull Parasite) lowers to a permanent target excluding lands.
func TestLowerRemoveCounterNonlandTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Thrull Parasite",
		Layout:     "normal",
		TypeLine:   "Creature — Thrull",
		OracleText: "{T}, Pay 2 life: Remove a counter from target nonland permanent.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v", face.ActivatedAbilities)
	}
	content := face.ActivatedAbilities[0].Content
	remove := removeCounterFromMode(t, content)
	if remove.Amount != game.Fixed(1) || !remove.ChooseKind {
		t.Fatalf("remove = %#v", remove)
	}
	excluded := content.Modes[0].Targets[0].Selection.Val.ExcludedTypes
	if len(excluded) != 1 {
		t.Fatalf("excluded types = %#v", excluded)
	}
}

// TestLowerRemoveCounterUnsupportedPluralChosenKind proves the kind-unspecified
// plural form ("remove two counters from target permanent") fails closed: it has
// no single-choice resolution, so it must not lower to a one-kind removal.
func TestLowerRemoveCounterUnsupportedPluralChosenKind(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Twin Drain",
		Layout:     "normal",
		TypeLine:   "Creature — Horror",
		OracleText: "{T}: Remove two counters from target permanent.",
	})
}

// TestLowerRemoveCounterUnsupportedDynamicAmount proves a non-fixed removal
// amount ("remove X counters") fails closed rather than lowering. The
// kind-agnostic "remove all counters" mass form is handled separately by
// TestLowerRemoveAllCountersTargetSpell and is intentionally not exercised here.
func TestLowerRemoveCounterUnsupportedDynamicAmount(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Counter Eater",
		Layout:     "normal",
		TypeLine:   "Creature — Horror",
		OracleText: "{X}, {T}: Remove X counters from target permanent.",
	})
}

// removeCounterFromSelfMode pulls the single RemoveCounter primitive out of a
// no-target, one-instruction ability mode whose object is the ability's own
// source, failing the test on any other shape.
func removeCounterFromSelfMode(t *testing.T, content game.AbilityContent) game.RemoveCounter {
	t.Helper()
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %#v", content.Modes)
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	remove, ok := mode.Sequence[0].Primitive.(game.RemoveCounter)
	if !ok {
		t.Fatalf("primitive = %#v", mode.Sequence[0].Primitive)
	}
	if remove.Object != game.SourcePermanentReference() {
		t.Fatalf("object = %#v", remove.Object)
	}
	return remove
}

// TestLowerRemoveCounterSelfNamedKind proves the source/self-referenced removal
// "Remove a -1/-1 counter from this creature." (Magmaroth, the Hatchling cycle)
// lowers to a RemoveCounter on the ability's own source naming that exact kind.
func TestLowerRemoveCounterSelfNamedKind(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Magmaroth",
		Layout:     "normal",
		TypeLine:   "Creature — Hippo Beast",
		OracleText: "At the beginning of your upkeep, remove a -1/-1 counter from this creature.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v", face.TriggeredAbilities)
	}
	remove := removeCounterFromSelfMode(t, face.TriggeredAbilities[0].Content)
	if remove.Amount != game.Fixed(1) || remove.ChooseKind || remove.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveCounterSelfActivated proves the activated self-removal "Remove
// a -1/-1 counter from this creature." (the Hatchling cycle) lowers to a
// source-object RemoveCounter.
func TestLowerRemoveCounterSelfActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sturdy Hatchling",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Elemental",
		OracleText: "{2}, Sacrifice another artifact or creature: Remove a -1/-1 counter from this creature.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v", face.ActivatedAbilities)
	}
	remove := removeCounterFromSelfMode(t, face.ActivatedAbilities[0].Content)
	if remove.Amount != game.Fixed(1) || remove.ChooseKind || remove.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveCounterSelfUnsupportedAll proves the "remove all" self form
// ("remove all +1/+1 counters from this creature.", Blood Hound) fails closed:
// its byte-exact reconstruction never matches the "all" wording.
func TestLowerRemoveCounterSelfUnsupportedAll(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Blood Hound",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental Dog",
		OracleText: "{2}{R}: Remove all +1/+1 counters from this creature.",
	})
}

// TestLowerRemoveAllCountersTargetSpell proves the kind-agnostic mass form
// "Remove all counters from target permanent." (Vampire Hexmage's ability body,
// here as a sorcery) lowers to a RemoveCounter with AllKinds set and no amount
// or named kind, so the runtime clears every counter regardless of kind.
func TestLowerRemoveAllCountersTargetSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Hexmage Sorcery",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Remove all counters from target permanent.",
	})
	remove := removeCounterFromMode(t, face.SpellAbility.Val)
	if !remove.AllKinds || remove.ChooseKind ||
		remove.Amount != (game.Quantity{}) || remove.CounterKind != 0 {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveAllCountersSelfActivated proves the self/source mass form
// "Remove all counters from this permanent." lowers onto the ability's own
// source with AllKinds set.
func TestLowerRemoveAllCountersSelfActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Hexmage Self",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{2}, {T}: Remove all counters from this permanent.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v", face.ActivatedAbilities)
	}
	content := face.ActivatedAbilities[0].Content
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v", content)
	}
	remove, ok := content.Modes[0].Sequence[0].Primitive.(game.RemoveCounter)
	if !ok || !remove.AllKinds || remove.Object != game.SourcePermanentReference() {
		t.Fatalf("primitive = %#v", content.Modes[0].Sequence[0].Primitive)
	}
}

// removeCounterFromGroupMode pulls the single RemoveCounter primitive out of a
// no-target, one-instruction ability mode (the group-recipient form "remove a
// counter from each <group>"), failing the test on any other shape.
func removeCounterFromGroupMode(t *testing.T, content game.AbilityContent) game.RemoveCounter {
	t.Helper()
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %#v", content.Modes)
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	remove, ok := mode.Sequence[0].Primitive.(game.RemoveCounter)
	if !ok {
		t.Fatalf("primitive = %#v", mode.Sequence[0].Primitive)
	}
	return remove
}

// TestLowerRemoveCounterGroupControllerRestricted proves the group-recipient
// form "remove a -1/-1 counter from each creature you control" (Heartmender)
// lowers to a RemoveCounter with a controller-restricted battlefield Group and
// no single Object, so the runtime removes the named kind from every matching
// permanent.
func TestLowerRemoveCounterGroupControllerRestricted(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Heartmender",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Remove a -1/-1 counter from each creature you control.",
	})
	remove := removeCounterFromGroupMode(t, face.SpellAbility.Val)
	if !remove.Group.Valid() || remove.Object != (game.ObjectReference{}) {
		t.Fatalf("group/object = %#v", remove)
	}
	if remove.Amount != game.Fixed(1) || remove.ChooseKind || remove.AllKinds ||
		remove.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveCounterGroupPlural proves the group form supports a plural
// amount and an unrestricted (any-controller) group: "Remove two loyalty
// counters from each planeswalker." (Pestilent Haze) lowers to a RemoveCounter
// removing two loyalty counters from every planeswalker.
func TestLowerRemoveCounterGroupPlural(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pestilent Haze",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Remove two loyalty counters from each planeswalker.",
	})
	remove := removeCounterFromGroupMode(t, face.SpellAbility.Val)
	if !remove.Group.Valid() || remove.Object != (game.ObjectReference{}) {
		t.Fatalf("group/object = %#v", remove)
	}
	if remove.Amount != game.Fixed(2) || remove.ChooseKind || remove.AllKinds ||
		remove.CounterKind != counter.Loyalty {
		t.Fatalf("remove = %#v", remove)
	}
}

// TestLowerRemoveCounterGroupUnspecifiedKind proves the kind-unspecified group
// form "Remove a counter from each creature you control." fails closed: a group
// removal has no single controller-chosen kind to resolve.
func TestLowerRemoveCounterGroupUnspecifiedKind(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Unspecified Group",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Remove a counter from each creature you control.",
	})
}
