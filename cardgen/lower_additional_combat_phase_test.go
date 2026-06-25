package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAdditionalCombatPhaseAggravatedAssault proves that Aggravated
// Assault's "{3}{R}{R}: Untap all creatures you control. After this main phase,
// there is an additional combat phase followed by an additional main phase.
// Activate only as a sorcery." activated ability lowers to a two-instruction
// sequence: an untap of the controller's creatures followed by an AddExtraPhases
// primitive that queues an extra combat and main phase.
func TestLowerAdditionalCombatPhaseAggravatedAssault(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Aggravated Assault",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{1}{R}",
		OracleText: "{3}{R}{R}: Untap all creatures you control. After this main phase, there is an additional combat phase followed by an additional main phase. Activate only as a sorcery.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if len(ability.Content.Modes) != 1 {
		t.Fatalf("ability content modes = %#v, want one mode", ability.Content.Modes)
	}
	seq := ability.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", seq)
	}
	if _, ok := seq[0].Primitive.(game.Untap); !ok {
		t.Fatalf("first primitive = %T, want game.Untap", seq[0].Primitive)
	}
	extra, ok := seq[1].Primitive.(game.AddExtraPhases)
	if !ok {
		t.Fatalf("second primitive = %T, want game.AddExtraPhases", seq[1].Primitive)
	}
	if !extra.Combat || !extra.Main {
		t.Fatalf("extra phases = %#v, want Combat and Main", extra)
	}
}

// TestLowerAdditionalCombatPhaseCombatOnly proves the shorter "After this phase,
// there is an additional combat phase." wording (Aurelia, Combat Celebrant)
// lowers to an AddExtraPhases that queues only an extra combat phase.
func TestLowerAdditionalCombatPhaseCombatOnly(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Aurelia",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "After this phase, there is an additional combat phase.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	seq := face.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", seq)
	}
	extra, ok := seq[0].Primitive.(game.AddExtraPhases)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddExtraPhases", seq[0].Primitive)
	}
	if !extra.Combat || extra.Main {
		t.Fatalf("extra phases = %#v, want Combat only", extra)
	}
}

// TestLowerAdditionalCombatPhaseRaiyuu proves Raiyuu, Storm's Edge lowers its
// "untap it. If it's the first combat phase of the turn, there is an additional
// combat phase after this phase." triggered ability into a two-instruction
// sequence: an ungated untap of the triggering attacker, followed by an
// AddExtraPhases gated by the FirstCombatPhaseOfTurn condition. The trailing
// "after this phase" word order and the leading condition clause are the
// Raiyuu-specific wordings exercised here.
func TestLowerAdditionalCombatPhaseRaiyuu(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Raiyuu, Storm's Edge",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Samurai",
		ManaCost:   "{1}{R}{R}",
		OracleText: "First strike\nWhenever a Samurai or Warrior you control attacks alone, untap it. If it's the first combat phase of the turn, there is an additional combat phase after this phase.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", seq)
	}
	if _, ok := seq[0].Primitive.(game.Untap); !ok {
		t.Fatalf("first primitive = %T, want game.Untap", seq[0].Primitive)
	}
	if seq[0].Condition.Exists {
		t.Fatalf("untap must be ungated, got condition %#v", seq[0].Condition.Val)
	}
	extra, ok := seq[1].Primitive.(game.AddExtraPhases)
	if !ok {
		t.Fatalf("second primitive = %T, want game.AddExtraPhases", seq[1].Primitive)
	}
	if !extra.Combat || extra.Main {
		t.Fatalf("extra phases = %#v, want Combat only", extra)
	}
	if !seq[1].Condition.Exists {
		t.Fatal("additional combat phase must be gated by a condition")
	}
	if !seq[1].Condition.Val.Condition.Val.FirstCombatPhaseOfTurn {
		t.Fatalf("gate condition = %#v, want FirstCombatPhaseOfTurn", seq[1].Condition.Val)
	}
}
