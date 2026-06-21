package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerTriggerBodyResolutionCondition verifies that a triggered ability
// whose resolving body carries a state condition checked only on resolution
// ("Whenever X, EFFECT. If STATE, EFFECT2.") routes through the shared content
// lowering exactly as the same body lowers on a spell, rather than being
// rejected by the trigger-body preparation gate. The condition is not the
// trigger's intervening "if", so it stays in the body.
func TestLowerTriggerBodyResolutionCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, put a +1/+1 counter on this creature. If you control eight or more lands, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Errorf("event = %v, want EventPermanentEnteredBattlefield", ta.Trigger.Pattern.Event)
	}
	if len(ta.Content.Modes) != 1 {
		t.Fatalf("modes = %#v, want 1", ta.Content.Modes)
	}
	seq := ta.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.AddCounter); !ok {
		t.Errorf("instruction[0] = %#v, want AddCounter", seq[0].Primitive)
	}
	if _, ok := seq[1].Primitive.(game.Draw); !ok {
		t.Errorf("instruction[1] = %#v, want Draw", seq[1].Primitive)
	}
}

// TestLowerChosenTypeCastTrigger verifies that a chosen-type cast trigger
// ("Whenever you cast a creature spell of the chosen type, draw a card.")
// lowers to an EventSpellCast pattern whose CardSelection carries the runtime
// SubtypeChoiceSourceEntry predicate alongside the creature card type, so
// the trigger fires only for spells sharing the source's entry-time creature
// type. The full Vanquisher's Banner card (entry choice + chosen-type anthem +
// chosen-type cast trigger) lowers end to end.
func TestLowerChosenTypeCastTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Vanquisher's Banner",
		Layout:   "normal",
		TypeLine: "Artifact",
		OracleText: "As Vanquisher's Banner enters, choose a creature type.\n" +
			"Creatures you control of the chosen type get +1/+1.\n" +
			"Whenever you cast a creature spell of the chosen type, draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventSpellCast {
		t.Errorf("event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
	}
	selection := ta.Trigger.Pattern.CardSelection
	if selection.SubtypeChoice != game.SubtypeChoiceSourceEntry {
		t.Errorf("CardSelection.SubtypeChoice != SubtypeChoiceSourceEntry, want SourceEntry: %#v", selection)
	}
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Errorf("CardSelection.RequiredTypes = %#v, want [Creature]", selection.RequiredTypes)
	}
}

// ability whose resolving body is an optional "you may X. If you do, Y" flow
// routes through the shared ordered-effect-sequence path even when the parser
// flags only the leading effect optional (ability.Optional is false) rather
// than the whole ability. The body lowers exactly as it does on a spell.
func TestLowerTriggerBodyOptionalSequenceEffectMarked(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wizard",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "Whenever you cast an instant or sorcery spell, you may draw a card. If you do, discard a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventSpellCast {
		t.Errorf("event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
	}
	if ta.Optional {
		t.Error("trigger should fire unconditionally; only the body's first instruction is optional")
	}
	if len(ta.Content.Modes) != 1 {
		t.Fatalf("modes = %#v, want 1", ta.Content.Modes)
	}
	seq := ta.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.Draw); !ok {
		t.Errorf("instruction[0] = %#v, want Draw", seq[0].Primitive)
	}
	if !seq[0].Optional || seq[0].PublishResult != optionalIfYouDoResultKey {
		t.Errorf("instruction[0] = %#v, want optional with if-you-do publish", seq[0])
	}
	if _, ok := seq[1].Primitive.(game.Discard); !ok {
		t.Errorf("instruction[1] = %#v, want Discard", seq[1].Primitive)
	}
	if !seq[1].ResultGate.Exists ||
		seq[1].ResultGate.Val.Key != optionalIfYouDoResultKey ||
		seq[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Errorf("instruction[1].ResultGate = %#v, want succeeded gate on if-you-do", seq[1].ResultGate)
	}
}

// TestLowerTriggerBodyInterveningOptionalSequence verifies that a triggered
// ability whose body combines an intervening "if" condition with an optional
// "you may X. If you do, Y" resolving sequence routes correctly: the intervening
// condition gates the trigger and the residual "if you do" gate stays in the
// body for the shared ordered-effect-sequence path to consume.
func TestLowerTriggerBodyInterveningOptionalSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Looter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Rogue",
		OracleText: "Whenever this creature deals combat damage to a player, if you control an artifact, you may discard a card. If you do, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.InterveningIf != "if you control an artifact" {
		t.Errorf("InterveningIf = %q, want the intervening condition text", ta.Trigger.InterveningIf)
	}
	if !ta.Trigger.InterveningCondition.Exists {
		t.Error("InterveningCondition missing; the intervening 'if' must gate the trigger")
	}
	if ta.Optional {
		t.Error("trigger should fire unconditionally; only the body's first instruction is optional")
	}
	if len(ta.Content.Modes) != 1 {
		t.Fatalf("modes = %#v, want 1", ta.Content.Modes)
	}
	seq := ta.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.Discard); !ok {
		t.Errorf("instruction[0] = %#v, want Discard", seq[0].Primitive)
	}
	if !seq[0].Optional || seq[0].PublishResult != optionalIfYouDoResultKey {
		t.Errorf("instruction[0] = %#v, want optional with if-you-do publish", seq[0])
	}
	if _, ok := seq[1].Primitive.(game.Draw); !ok {
		t.Errorf("instruction[1] = %#v, want Draw", seq[1].Primitive)
	}
	if !seq[1].ResultGate.Exists ||
		seq[1].ResultGate.Val.Key != optionalIfYouDoResultKey ||
		seq[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Errorf("instruction[1].ResultGate = %#v, want succeeded gate on if-you-do", seq[1].ResultGate)
	}
}
