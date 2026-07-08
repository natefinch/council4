package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestGenerateBlitzwingCruelTormentorFront proves the front face of Blitzwing,
// Cruel Tormentor lowers its end-step trigger to a target-opponent life loss
// scaled by that opponent's own life-lost-this-turn (gap #1), whose loss result
// gates a convert-self that fires only when no life was lost (gap #2). The
// dynamic amount reads the target player (not the controller), and the convert
// is gated on Succeeded=TriFalse.
func TestGenerateBlitzwingCruelTormentorFront(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Blitzwing, Cruel Tormentor",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact Creature — Robot",
		ManaCost:  "{5}{B}",
		Power:     new("6"),
		Toughness: new("5"),
		OracleText: "More Than Meets the Eye {3}{B} (You may cast this card converted for {3}{B}.)\n" +
			"At the beginning of your end step, target opponent loses life equal to the life that player lost this turn. If no life is lost this way, convert Blitzwing.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		ability.Trigger.Pattern.Step != game.StepEnd ||
		ability.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger pattern = %#v, want your end step", ability.Trigger.Pattern)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence len = %d, want 2", len(mode.Sequence))
	}

	loseLife, ok := mode.Sequence[0].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.LoseLife", mode.Sequence[0].Primitive)
	}
	if loseLife.Player != game.TargetPlayerReference(0) {
		t.Fatalf("lose-life recipient = %#v, want TargetPlayerReference(0)", loseLife.Player)
	}
	dynamic := loseLife.Amount.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("lose-life amount = %#v, want dynamic", loseLife.Amount)
	}
	if dynamic.Val.Kind != game.DynamicAmountLifeLostThisTurn {
		t.Fatalf("dynamic kind = %v, want DynamicAmountLifeLostThisTurn", dynamic.Val.Kind)
	}
	if dynamic.Val.Player == nil || *dynamic.Val.Player != game.TargetPlayerReference(0) {
		t.Fatalf("dynamic player = %#v, want TargetPlayerReference(0)", dynamic.Val.Player)
	}
	if mode.Sequence[0].PublishResult == "" {
		t.Fatal("lose-life instruction does not publish its loss result")
	}

	transform, ok := mode.Sequence[1].Primitive.(game.Transform)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.Transform", mode.Sequence[1].Primitive)
	}
	if transform.Object != game.SourcePermanentReference() {
		t.Fatalf("transform object = %#v, want SourcePermanentReference()", transform.Object)
	}
	gate := mode.Sequence[1].ResultGate
	if !gate.Exists {
		t.Fatal("transform instruction has no result gate")
	}
	if gate.Val.Key != mode.Sequence[0].PublishResult {
		t.Fatalf("gate key = %q, want %q", gate.Val.Key, mode.Sequence[0].PublishResult)
	}
	if gate.Val.Succeeded != game.TriFalse {
		t.Fatalf("gate Succeeded = %v, want TriFalse (convert only when no life lost)", gate.Val.Succeeded)
	}
}

// TestGenerateBlitzwingAdaptiveAssailantBack proves the back face's
// beginning-of-combat trigger lowers the two-sentence "choose flying or
// indestructible at random. Blitzwing gains that ability until end of turn."
// construction (gap #3) into an at-random one-of-two modal keyword grant: each
// mode grants one of the two keywords to the source until end of turn, exactly
// one mode is required, and RandomModes selects it with the game's random source.
func TestGenerateBlitzwingAdaptiveAssailantBack(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Blitzwing, Adaptive Assailant",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact — Vehicle",
		Power:     new("3"),
		Toughness: new("5"),
		OracleText: "Living metal (During your turn, this Vehicle is also a creature.)\n" +
			"At the beginning of combat on your turn, choose flying or indestructible at random. Blitzwing gains that ability until end of turn.\n" +
			"Whenever Blitzwing deals combat damage to a player, convert it.",
	})
	var combat game.TriggeredAbility
	found := false
	for _, ability := range face.TriggeredAbilities {
		if ability.Trigger.Pattern.Step == game.StepBeginningOfCombat {
			combat = ability
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no beginning-of-combat trigger in %#v", face.TriggeredAbilities)
	}
	if combat.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		combat.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("combat trigger pattern = %#v, want beginning of your combat", combat.Trigger.Pattern)
	}
	content := combat.Content
	if !content.RandomModes {
		t.Fatal("combat trigger content is not RandomModes")
	}
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("mode range = [%d,%d], want exactly one", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 2 {
		t.Fatalf("modes = %d, want 2 (flying, indestructible)", len(content.Modes))
	}
	want := []game.Keyword{game.Flying, game.Indestructible}
	for i, mode := range content.Modes {
		if len(mode.Sequence) != 1 {
			t.Fatalf("mode[%d] sequence len = %d, want 1", i, len(mode.Sequence))
		}
		apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
		if !ok {
			t.Fatalf("mode[%d] primitive = %T, want game.ApplyContinuous", i, mode.Sequence[0].Primitive)
		}
		if !apply.Object.Exists || apply.Object.Val != game.SourceCardPermanentReference() {
			t.Fatalf("mode[%d] object = %#v, want SourceCardPermanentReference()", i, apply.Object)
		}
		if apply.Duration != game.DurationUntilEndOfTurn {
			t.Fatalf("mode[%d] duration = %v, want DurationUntilEndOfTurn", i, apply.Duration)
		}
		if len(apply.ContinuousEffects) != 1 {
			t.Fatalf("mode[%d] continuous effects = %d, want 1", i, len(apply.ContinuousEffects))
		}
		effect := apply.ContinuousEffects[0]
		if effect.Layer != game.LayerAbility {
			t.Fatalf("mode[%d] layer = %v, want LayerAbility", i, effect.Layer)
		}
		if len(effect.AddKeywords) != 1 || effect.AddKeywords[0] != want[i] {
			t.Fatalf("mode[%d] keywords = %v, want [%v]", i, effect.AddKeywords, want[i])
		}
	}
}
