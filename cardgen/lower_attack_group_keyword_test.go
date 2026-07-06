package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerAttackTriggerGroupKeywordGrant verifies that "Whenever one or more
// creatures you control attack, they gain <keyword> until end of turn."
// (Angelic Guardian) lowers to a triggered ability whose effect grants the
// keyword to the attacking-creatures-you-control battlefield group via a single
// ApplyContinuous until end of turn.
func TestLowerAttackTriggerGroupKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Angelic Guardian",
		Layout:     "normal",
		ManaCost:   "{4}{W}",
		TypeLine:   "Creature — Angel",
		OracleText: "Flying\nWhenever one or more creatures you control attack, they gain indestructible until end of turn.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	triggered := face.TriggeredAbilities[0]
	pattern := triggered.Trigger.Pattern
	if pattern.Event != game.EventAttackerDeclared ||
		pattern.Controller != game.TriggerControllerYou ||
		!pattern.OneOrMore {
		t.Fatalf("trigger pattern = %+v, want one-or-more controlled-creature attack", pattern)
	}
	if triggered.Optional {
		t.Fatal("trigger must be mandatory")
	}
	mode := triggered.Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want 0", len(mode.Targets))
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single ApplyContinuous", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if apply.Object.Exists || apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply = %+v, want unanchored group grant until end of turn", apply)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(apply.ContinuousEffects))
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility {
		t.Fatalf("layer = %v, want LayerAbility", effect.Layer)
	}
	if len(effect.AddKeywords) != 1 || effect.AddKeywords[0] != game.Indestructible {
		t.Fatalf("keywords = %v, want [Indestructible]", effect.AddKeywords)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerYou ||
		selection.CombatState != game.CombatStateAttacking ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want attacking creatures you control", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("attacking creatures you control must not exclude the source")
	}
}

// TestLowerAttackTriggerGroupKeywordGrantFailsClosed verifies the attack-group
// keyword-grant path does not fire for shapes it must not handle, so those cards
// still fail closed rather than being mis-lowered. A parameterized keyword grant
// ("they gain annihilator 1") is not a plain keyword grant and must be rejected.
func TestLowerAttackTriggerGroupKeywordGrantFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Attack Group Parameterized",
		Layout:     "normal",
		ManaCost:   "{4}{W}",
		TypeLine:   "Creature — Angel",
		OracleText: "Whenever one or more creatures you control attack, they gain annihilator 1 until end of turn.",
	})
}
