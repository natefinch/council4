package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerEndlessEvilAttachedDiesReturnsSource proves the full Endless Evil
// lowering: the "When enchanted creature dies, if that creature was a Horror,
// return this card to its owner's hand." ability lowers to an attached-permanent
// dies trigger whose intervening-if gate matches the dead creature's Horror
// subtype via last-known information, and whose body returns the Aura source
// permanent to its owner's hand. The upkeep ability lowers to a 1/1 copy of the
// enchanted creature.
func TestLowerEndlessEvilAttachedDiesReturnsSource(t *testing.T) {
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Endless Evil",
		Layout:   "normal",
		TypeLine: "Enchantment — Aura",
		OracleText: "Enchant creature you control\n" +
			"At the beginning of your upkeep, create a token that's a copy of enchanted creature, except the token is 1/1.\n" +
			"When enchanted creature dies, if that creature was a Horror, return this card to its owner's hand.",
	})

	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}

	// Upkeep 1/1 copy token.
	upkeep := face.TriggeredAbilities[0]
	if upkeep.Trigger.Pattern.Event != game.EventBeginningOfStep || upkeep.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("upkeep trigger pattern = %+v, want beginning of upkeep", upkeep.Trigger.Pattern)
	}
	create, ok := upkeep.Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("upkeep primitive = %T, want CreateToken", upkeep.Content.Modes[0].Sequence[0].Primitive)
	}
	spec, ok := create.Source.TokenCopy()
	if !ok {
		t.Fatal("upkeep token source is not a copy spec")
	}
	if spec.Object.Kind() != game.ObjectReferenceSourceAttachedPermanent {
		t.Fatalf("copy source object = %v, want the attached (enchanted) permanent", spec.Object.Kind())
	}
	if !spec.SetPower.Exists || spec.SetPower.Val.Value != 1 || !spec.SetToughness.Exists || spec.SetToughness.Val.Value != 1 {
		t.Fatalf("copy override P/T = %v/%v, want 1/1", spec.SetPower, spec.SetToughness)
	}

	// Attached-permanent dies trigger with LKI Horror gate returning the source.
	dies := face.TriggeredAbilities[1]
	if dies.Trigger.Type != game.TriggerWhen {
		t.Fatalf("dies trigger type = %v, want TriggerWhen", dies.Trigger.Type)
	}
	if dies.Trigger.Pattern.Event != game.EventPermanentDied {
		t.Fatalf("dies trigger event = %v, want EventPermanentDied", dies.Trigger.Pattern.Event)
	}
	if dies.Trigger.Pattern.Source != game.TriggerSourceAttachedPermanent {
		t.Fatalf("dies trigger source = %v, want TriggerSourceAttachedPermanent", dies.Trigger.Pattern.Source)
	}
	if !dies.Trigger.InterveningCondition.Exists {
		t.Fatal("dies trigger has no intervening condition")
	}
	cond := dies.Trigger.InterveningCondition.Val
	if !cond.Object.Exists || cond.Object.Val.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("intervening condition object = %+v, want the event permanent", cond.Object)
	}
	if !cond.ObjectMatches.Exists ||
		len(cond.ObjectMatches.Val.SubtypesAny) != 1 ||
		cond.ObjectMatches.Val.SubtypesAny[0] != types.Horror {
		t.Fatalf("intervening condition selection = %+v, want Horror subtype", cond.ObjectMatches)
	}
	move, ok := dies.Content.Modes[0].Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("dies body primitive = %T, want MoveCard", dies.Content.Modes[0].Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceSource {
		t.Fatalf("move card reference = %v, want the source card", move.Card.Kind)
	}
	if move.FromZone != zone.Graveyard {
		t.Fatalf("move from zone = %v, want graveyard", move.FromZone)
	}
	if move.Destination != zone.Hand {
		t.Fatalf("move destination = %v, want hand", move.Destination)
	}
}

// TestLowerAttachedDiesNonHorrorSubtypeGate proves the "that <type> was a
// <subtype>" intervening gate is reusable typed support, not card-name logic: an
// otherwise identical Aura keyed on a different subtype lowers to the same shape
// with that subtype in the gate.
func TestLowerAttachedDiesNonHorrorSubtypeGate(t *testing.T) {
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Aura",
		Layout:   "normal",
		TypeLine: "Enchantment — Aura",
		OracleText: "Enchant creature\n" +
			"When enchanted creature dies, if that creature was a Zombie, return this card to its owner's hand.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	cond := face.TriggeredAbilities[0].Trigger.InterveningCondition
	if !cond.Exists || !cond.Val.ObjectMatches.Exists ||
		len(cond.Val.ObjectMatches.Val.SubtypesAny) != 1 ||
		cond.Val.ObjectMatches.Val.SubtypesAny[0] != types.Zombie {
		t.Fatalf("intervening condition = %+v, want Zombie subtype", cond)
	}
}
