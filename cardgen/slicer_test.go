package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestGenerateSlicerHiredMuscleFront proves the front face lowers its
// "At the beginning of each opponent's upkeep" trigger into the optional
// give-control-of-the-source-to-the-triggering-opponent-until-end-of-turn offer
// (mechanic: a LayerControl continuous effect whose new controller binds to the
// triggering event's player, wrapped in the "you may" optional and publishing
// its result), followed by the "If you do" gated untap, goad, and
// can't-be-sacrificed-this-turn consequences and the "If you don't" gated
// self-convert else branch.
func TestGenerateSlicerHiredMuscleFront(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Slicer, Hired Muscle",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact Creature — Robot",
		ManaCost:  "{4}{R}",
		Power:     new("3"),
		Toughness: new("4"),
		OracleText: "More Than Meets the Eye {2}{R} (You may cast this card converted for {2}{R}.)\n" +
			"Double strike, haste\n" +
			"At the beginning of each opponent's upkeep, you may have that player gain control of Slicer until end of turn. If you do, untap Slicer, goad it, and it can't be sacrificed this turn. If you don't, convert it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Type != game.TriggerAt ||
		ability.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		ability.Trigger.Pattern.Controller != game.TriggerControllerOpponent ||
		ability.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("trigger = %#v, want at each opponent's upkeep", ability.Trigger)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 5 {
		t.Fatalf("sequence len = %d, want 5", len(mode.Sequence))
	}

	// Mechanic: optional give-control of the source to the triggering opponent.
	give := mode.Sequence[0]
	apply, ok := give.Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.ApplyContinuous", give.Primitive)
	}
	if apply.Object != opt.Val(game.SourcePermanentReference()) {
		t.Fatalf("give object = %#v, want SourcePermanentReference()", apply.Object)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("give duration = %v, want DurationUntilEndOfTurn", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("give continuous effects = %d, want 1", len(apply.ContinuousEffects))
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerControl {
		t.Fatalf("give layer = %v, want LayerControl", effect.Layer)
	}
	if effect.NewControllerRef != opt.Val(game.EventPlayerReference()) {
		t.Fatalf("give new controller = %#v, want EventPlayerReference()", effect.NewControllerRef)
	}
	if !give.Optional {
		t.Fatal("give instruction is not optional (the \"you may\")")
	}
	if give.PublishResult == "" {
		t.Fatal("give instruction does not publish its result for the \"If you do\"/\"If you don't\" gates")
	}

	// Mechanic: "If you do" untap Slicer.
	untap := mode.Sequence[1]
	if _, ok := untap.Primitive.(game.Untap); !ok {
		t.Fatalf("sequence[1] = %T, want game.Untap", untap.Primitive)
	}
	assertGatedOn(t, untap, give.PublishResult, game.TriTrue, "untap")

	// Mechanic: "If you do" goad Slicer.
	goad := mode.Sequence[2]
	if _, ok := goad.Primitive.(game.Goad); !ok {
		t.Fatalf("sequence[2] = %T, want game.Goad", goad.Primitive)
	}
	assertGatedOn(t, goad, give.PublishResult, game.TriTrue, "goad")

	// Mechanic: "If you do" it can't be sacrificed this turn.
	shield := mode.Sequence[3]
	rule, ok := shield.Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("sequence[3] = %T, want game.ApplyRule", shield.Primitive)
	}
	if rule.Duration != game.DurationThisTurn {
		t.Fatalf("shield duration = %v, want DurationThisTurn", rule.Duration)
	}
	if len(rule.RuleEffects) != 1 ||
		rule.RuleEffects[0].Kind != game.RuleEffectCantBeSacrificed ||
		!rule.RuleEffects[0].AffectedSource {
		t.Fatalf("shield rule effects = %#v, want [{CantBeSacrificed AffectedSource}]", rule.RuleEffects)
	}
	assertGatedOn(t, shield, give.PublishResult, game.TriTrue, "can't-be-sacrificed")

	// Mechanic: "If you don't" convert it (else branch).
	convert := mode.Sequence[4]
	transform, ok := convert.Primitive.(game.Transform)
	if !ok {
		t.Fatalf("sequence[4] = %T, want game.Transform", convert.Primitive)
	}
	if transform.Object != game.SourcePermanentReference() {
		t.Fatalf("convert object = %#v, want SourcePermanentReference()", transform.Object)
	}
	assertGatedOn(t, convert, give.PublishResult, game.TriFalse, "convert")
}

// TestGenerateSlicerHighSpeedAntagonistBack proves the back face lowers its
// "Whenever Slicer deals combat damage to a player" trigger into the delayed
// convert-at-end-of-combat self-transform: a CreateDelayedTrigger gated on the
// combat-damage event scheduling a DelayedAtEndOfCombat transform of the source.
func TestGenerateSlicerHighSpeedAntagonistBack(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Slicer, High-Speed Antagonist",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact — Vehicle",
		Power:     new("3"),
		Toughness: new("2"),
		OracleText: "Living metal (During your turn, this Vehicle is also a creature.)\n" +
			"First strike, haste\n" +
			"Whenever Slicer deals combat damage to a player, convert it at end of combat.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Type != game.TriggerWhenever ||
		ability.Trigger.Pattern.Event != game.EventDamageDealt ||
		ability.Trigger.Pattern.Source != game.TriggerSourceSelf ||
		!ability.Trigger.Pattern.RequireCombatDamage ||
		ability.Trigger.Pattern.DamageRecipient != game.DamageRecipientPlayer {
		t.Fatalf("trigger = %#v, want whenever this deals combat damage to a player", ability.Trigger)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	delayed, ok := mode.Sequence[0].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.CreateDelayedTrigger", mode.Sequence[0].Primitive)
	}
	if delayed.Trigger.Timing != game.DelayedAtEndOfCombat {
		t.Fatalf("delayed timing = %v, want DelayedAtEndOfCombat", delayed.Trigger.Timing)
	}
	inner := delayed.Trigger.Content.Modes[0]
	if len(inner.Sequence) != 1 {
		t.Fatalf("delayed sequence len = %d, want 1", len(inner.Sequence))
	}
	transform, ok := inner.Sequence[0].Primitive.(game.Transform)
	if !ok {
		t.Fatalf("delayed primitive = %T, want game.Transform", inner.Sequence[0].Primitive)
	}
	if transform.Object != game.SourceCardPermanentReference() {
		t.Fatalf("delayed transform object = %#v, want SourceCardPermanentReference()", transform.Object)
	}
}

func assertGatedOn(t *testing.T, in game.Instruction, key game.ResultKey, want game.TriState, label string) {
	t.Helper()
	if !in.ResultGate.Exists {
		t.Fatalf("%s has no result gate", label)
	}
	if in.ResultGate.Val.Key != key {
		t.Fatalf("%s gate key = %q, want %q", label, in.ResultGate.Val.Key, key)
	}
	if in.ResultGate.Val.Succeeded != want {
		t.Fatalf("%s gate Succeeded = %v, want %v", label, in.ResultGate.Val.Succeeded, want)
	}
}
