package parser

import "testing"

// counterPlacementExact parses a single counter-placement sentence and reports
// whether its resolving effect round-tripped to an exact, lowerable production.
func counterPlacementExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectPut {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactStunCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a stun counter on target creature.",
		"Put a stun counter on target creature an opponent controls.",
		"Put two stun counters on target creature.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

func TestExactFinalityCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	// Finality counters have no complete runtime semantics, so their placement
	// clause stays inexact and unlowerable.
	rejected := []string{
		"Put a finality counter on target creature.",
		"Put two finality counters on target creature.",
	}
	for _, source := range rejected {
		if counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = true, want false", source)
		}
	}
}

func TestExactMultiTargetCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on each of up to two target creatures.",
		"Put a +1/+1 counter on each of up to three target creatures.",
		"Put a +1/+1 counter on each of up to two target creatures you control.",
		"Put a +1/+1 counter on each of up to two other target creatures.",
		"Put a +1/+1 counter on each of up to two other target creatures you control.",
		"Put a -1/-1 counter on each of up to two target creatures.",
		// The unbounded "each of any number of target" multi-target object is
		// reconstructed exactly; the lowering layer decides which forms it models.
		"Put a +1/+1 counter on each of any number of target creatures.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

func TestExactMultiTargetCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// The "each of" coordinator is required for the multi-target object.
		"Put a +1/+1 counter on up to two target creatures.",
		// Subtype-restricted plural targets are not a plain permanent noun.
		"Put a +1/+1 counter on each of up to two target Merfolk.",
	}
	for _, source := range rejected {
		if counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = true, want false", source)
		}
	}
}

// TestExactCounterPlacementControllerKeywordOrderingAccepts covers single-target
// recipients whose controller clause precedes a "with"/"without" keyword or a
// numeric "with power/toughness" qualifier, matching the canonical Oracle word
// order ("target creature you control without flying").
func TestExactCounterPlacementControllerKeywordOrderingAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on target creature you control without flying.",
		"Put a +1/+1 counter on target creature you control with flying.",
		"Put a +1/+1 counter on target creature you don't control with flying.",
		"Put a +1/+1 counter on target creature an opponent controls with flying.",
		"Put a +1/+1 counter on target creature you control with power 2.",
		"Put a +1/+1 counter on another target creature you control without flying.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactCounterPlacementGroupConjunctiveTypeAccepts covers group recipients
// whose noun conjoins two card types the permanent must carry at once ("each
// artifact creature you control") — Steel Overseer. The conjunctive marker keeps
// the two-word type a single all-of filter rather than an any-of union.
func TestExactCounterPlacementGroupConjunctiveTypeAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on each artifact creature you control.",
		"Put a +1/+1 counter on each enchantment creature you control.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
		effect := counterPlacementEffect(t, source)
		if !effect.Selection.ConjunctiveTypes {
			t.Errorf("Selection.ConjunctiveTypes(%q) = false, want true", source)
		}
		if len(effect.Selection.RequiredTypesAny) != 2 {
			t.Errorf("Selection.RequiredTypesAny(%q) = %v, want two types", source, effect.Selection.RequiredTypesAny)
		}
	}
}

// TestExactCounterPlacementGroupControllerKeywordOrderingAccepts covers group
// recipients whose controller clause precedes a keyword qualifier ("each
// creature you control with flying"), the dominant Oracle ordering.
func TestExactCounterPlacementGroupControllerKeywordOrderingAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on each creature you control with flying.",
		"Put a +1/+1 counter on each creature you control without flying.",
		"Put a +1/+1 counter on each creature you control with menace.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactCounterPlacementGroupEnteredThisTurnAccepts covers the "that entered
// this turn" temporal group filter (Oran-Rief, the Vastwood; Raucous
// Entertainer), asserting the recipient round-trips byte-exactly and the parser
// records the EnteredThisTurn flag on the recipient selection.
func TestExactCounterPlacementGroupEnteredThisTurnAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		controller SelectionController
	}{
		{"Put a +1/+1 counter on each green creature that entered this turn.", SelectionControllerAny},
		{"Put a +1/+1 counter on each creature you control that entered this turn.", SelectionControllerYou},
		{"Put a +1/+1 counter on each creature that entered this turn.", SelectionControllerAny},
	}
	for _, test := range tests {
		if !counterPlacementExact(t, test.source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", test.source)
			continue
		}
		effect := counterPlacementEffect(t, test.source)
		if !effect.Selection.EnteredThisTurn {
			t.Errorf("Parse(%q) selection.EnteredThisTurn = false, want true", test.source)
		}
		if effect.Selection.Controller != test.controller {
			t.Errorf("Parse(%q) selection.Controller = %v, want %v", test.source, effect.Selection.Controller, test.controller)
		}
	}
}

// TestExactCounterPlacementGroupEachOpponentControlsAccepts covers the
// distributive opponent group recipient ("each creature each opponent controls",
// Aku Djinn). It denotes the same opponent-controlled group as the plural "your
// opponents control" wording but rebuilds the verbatim distributive phrasing, so
// the parser records SelectionControllerOpponent with the OpponentEach marker.
func TestExactCounterPlacementGroupEachOpponentControlsAccepts(t *testing.T) {
	t.Parallel()
	const source = "Put a +1/+1 counter on each creature each opponent controls."
	if !counterPlacementExact(t, source) {
		t.Fatalf("counterPlacementExact(%q) = false, want true", source)
	}
	effect := counterPlacementEffect(t, source)
	if effect.Selection.Controller != SelectionControllerOpponent {
		t.Errorf("Selection.Controller = %v, want SelectionControllerOpponent", effect.Selection.Controller)
	}
	if !effect.Selection.OpponentEach {
		t.Error("Selection.OpponentEach = false, want true")
	}
	// The plural opponent wording denotes the same group but must not set the
	// distributive marker, so the two phrasings rebuild byte-exactly.
	plural := counterPlacementEffect(t, "Put a +1/+1 counter on each creature your opponents control.")
	if plural.Selection.Controller != SelectionControllerOpponent || plural.Selection.OpponentEach {
		t.Errorf("plural opponent wording = (%v, OpponentEach=%v), want (Opponent, false)", plural.Selection.Controller, plural.Selection.OpponentEach)
	}
}

// TestExactCounterPlacementGroupDynamicWhereXAccepts covers dynamic-X group
// counter placements whose "where X is <dynamic>" count phrase trails the
// recipient group (Ouroboroid, Southern Air Temple). The count subject's
// referent or subtype must not fold into the recipient, which stays the plain
// "each creature you control" group, so the recipient round-trips byte-exactly.
func TestExactCounterPlacementGroupDynamicWhereXAccepts(t *testing.T) {
	t.Parallel()
	sources := []string{
		"Put X +1/+1 counters on each creature you control, where X is this creature's power.",
		"Put X +1/+1 counters on each creature you control, where X is the number of Shrines you control.",
	}
	for _, source := range sources {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
			continue
		}
		effect := counterPlacementEffect(t, source)
		if effect.Selection.Kind != SelectionCreature || effect.Selection.Controller != SelectionControllerYou {
			t.Errorf("Parse(%q) recipient = (%v, %v), want (creature, you control)", source, effect.Selection.Kind, effect.Selection.Controller)
		}
		if len(effect.Selection.SubtypesAny) != 0 {
			t.Errorf("Parse(%q) recipient subtypes = %v, want none", source, effect.Selection.SubtypesAny)
		}
	}
}

// counterPlacementEffect parses a single counter-placement sentence and returns
// its resolving effect for recipient-shape assertions.
func counterPlacementEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectPut {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

// TestExactAttachedCounterPlacementAccepts covers the Aura recipient "enchanted
// creature": the counter is placed on the permanent the source is attached to.
func TestExactAttachedCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on enchanted creature.",
		"Put two -1/-1 counters on enchanted creature.",
		"Put six +1/+1 counters on enchanted creature.",
	}
	for _, source := range accepted {
		effect := counterPlacementEffect(t, source)
		if !effect.CounterRecipientAttached {
			t.Errorf("CounterRecipientAttached(%q) = false, want true", source)
		}
		if !effect.Exact {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactAttachedCounterPlacementFailsClosed keeps recipients that are not the
// bare "enchanted creature" out of the attached-recipient form.
func TestExactAttachedCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// A trailing selector qualifier is not the bare recipient.
		"Put a +1/+1 counter on enchanted creature with flying.",
		// "enchanted permanent" is a different recipient the runtime is not asked
		// to model here.
		"Put a +1/+1 counter on enchanted permanent.",
	}
	for _, source := range rejected {
		effect := counterPlacementEffect(t, source)
		if effect.CounterRecipientAttached && effect.Exact {
			t.Errorf("counterPlacement(%q) accepted as exact attached recipient, want fail-closed", source)
		}
	}
}

// TestThatManyCounterPlacementAmount proves that "put that many <kind> counters"
// records a generic triggering-event amount rather than misreading the "+1" of
// "+1/+1" as a fixed amount of one. The trigger that supplies the quantity is
// resolved later in lowering.
func TestThatManyCounterPlacementAmount(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Put that many +1/+1 counters on this creature.",
		"Put that many +1/+1 counters on it.",
		"Put that many charge counters on this artifact.",
	} {
		effect := counterPlacementEffect(t, source)
		if effect.Amount.Known {
			t.Errorf("counterPlacement(%q) amount Known = true, want unknown dynamic", source)
		}
		if effect.Amount.DynamicKind != EffectDynamicAmountTriggeringEventAmount {
			t.Errorf("counterPlacement(%q) DynamicKind = %q, want %q",
				source, effect.Amount.DynamicKind, EffectDynamicAmountTriggeringEventAmount)
		}
	}
}

// TestExactEquippedCounterPlacementAccepts covers the Equipment recipient
// "equipped creature", which shares the attached-permanent reference with the
// Aura "enchanted creature" recipient and so must round-trip exactly too.
func TestExactEquippedCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on equipped creature.",
		"Put a charge counter on equipped creature.",
		"Put two +1/+1 counters on equipped creature.",
	}
	for _, source := range accepted {
		effect := counterPlacementEffect(t, source)
		if !effect.CounterRecipientAttached {
			t.Errorf("CounterRecipientAttached(%q) = false, want true", source)
		}
		if !effect.Exact {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactForEachCounterPlacementAccepts covers the "for each <group>" dynamic
// count form, which places one counter per counted object on the source.
func TestExactForEachCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on this creature for each creature card in your graveyard.",
		"Put a +1/+1 counter on this creature for each card in your hand.",
		"Put a +1/+1 counter on this creature for each creature you control.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactForEachCounterPlacementFailsClosed keeps for-each forms the runtime
// cannot model out of the exact production: a count word above one (the bare "a
// <kind> counter" wording is required) and an iterator the dynamic-amount lowerer
// does not recognize both leave the effect inexact.
func TestExactForEachCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// A multiplier above one is not the bare "a <kind> counter" count word.
		"Put two +1/+1 counters on this creature for each creature you control.",
		// Combat-state iterators are not recognized dynamic count subjects.
		"Put a +1/+1 counter on this creature for each attacking creature you control.",
	}
	for _, source := range rejected {
		if counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = true, want false", source)
		}
	}
}

// TestExactTargetPlayerControlledGroupCounterPlacementAccepts keeps the group
// counter placement whose recipient is every permanent a single targeted player
// or opponent controls inside the exact production. The targeted player supplies
// the recipient group's controller relationship rather than receiving the
// counter, so the recipient group and its targeted player reconstruct jointly.
func TestExactTargetPlayerControlledGroupCounterPlacementAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a +1/+1 counter on each creature target player controls.",
		"Put a -1/-1 counter on each creature target player controls.",
		"Put a +1/+1 counter on each creature target opponent controls.",
		"Put two +1/+1 counters on each creature target player controls.",
		"Put a +1/+1 counter on each artifact target player controls.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}

// TestExactTargetPlayerCounterPlacementStillAccepts is a regression guard that
// the joint "<group> target player controls" reconstruction did not capture the
// ordinary single-recipient placements: a counter placed on the targeted player
// itself ("Put a poison counter on target player.") and a counter placed on a
// single targeted creature both still round-trip through their own exact
// productions.
func TestExactTargetPlayerCounterPlacementStillAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Put a poison counter on target player.",
		"Put a +1/+1 counter on target creature.",
		"Put a +1/+1 counter on each creature you control.",
	}
	for _, source := range accepted {
		if !counterPlacementExact(t, source) {
			t.Errorf("counterPlacementExact(%q) = false, want true", source)
		}
	}
}
