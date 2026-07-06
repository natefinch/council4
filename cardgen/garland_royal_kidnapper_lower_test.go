package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// garlandRoyalKidnapperCard is the Scryfall input for Garland, Royal Kidnapper.
func garlandRoyalKidnapperCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Garland, Royal Kidnapper",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Human Knight",
		ManaCost: "{2}{U}{B}",
		OracleText: "When Garland enters, target opponent becomes the monarch.\n" +
			"Whenever an opponent becomes the monarch, gain control of target creature that player controls for as long as they're the monarch.\n" +
			"Creatures you control but don't own get +2/+2 and can't be sacrificed.",
	}
}

// TestLowerGarlandGainControlWhileMonarchAbility proves Garland's second ability
// lowers to an ApplyContinuous gain-control instruction whose duration ends when
// the triggering opponent stops being the monarch, whose control expiry binds to
// the event player, and whose target is restricted to a creature that opponent
// controls.
func TestLowerGarlandGainControlWhileMonarchAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, garlandRoyalKidnapperCard())

	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	gainControl := face.TriggeredAbilities[1]
	if gainControl.Trigger.Pattern.Event != game.EventBecameMonarch ||
		gainControl.Trigger.Pattern.Player != game.TriggerPlayerOpponent {
		t.Fatalf("trigger pattern = %#v, want an opponent becoming the monarch", gainControl.Trigger.Pattern)
	}

	mode := gainControl.Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	sel := mode.Targets[0].Selection
	if !sel.Exists || !sel.Val.ControlledByEventPlayer {
		t.Fatalf("target selection = %#v, want ControlledByEventPlayer", mode.Targets[0].Selection)
	}
	if len(sel.Val.RequiredTypesAny) != 1 || sel.Val.RequiredTypesAny[0] != types.Creature {
		t.Fatalf("target required types = %#v, want any creature", sel.Val.RequiredTypesAny)
	}

	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %#v, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationForAsLongAsPlayerIsMonarch {
		t.Fatalf("duration = %v, want DurationForAsLongAsPlayerIsMonarch", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(apply.ContinuousEffects))
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerControl {
		t.Fatalf("effect layer = %v, want LayerControl", effect.Layer)
	}
	if !effect.NewController.Exists || effect.NewController.Val != game.Player1 {
		t.Fatalf("new controller = %#v, want the resolving controller (Player1)", effect.NewController)
	}
	if !effect.ExpiresForRef.Exists || effect.ExpiresForRef.Val.Kind() != game.PlayerReferenceEventPlayer {
		t.Fatalf("expires-for ref = %#v, want the event player", effect.ExpiresForRef)
	}
}

// TestLowerGarlandControlNotOwnStaticAbility proves Garland's third ability
// lowers to a +2/+2 anthem plus a can't-be-sacrificed rule effect, both scoped
// to creatures the controller controls but does not own.
func TestLowerGarlandControlNotOwnStaticAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, garlandRoyalKidnapperCard())

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body

	if len(body.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(body.ContinuousEffects))
	}
	anthem := body.ContinuousEffects[0]
	if anthem.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("anthem layer = %v, want LayerPowerToughnessModify", anthem.Layer)
	}
	if anthem.PowerDelta != 2 || anthem.ToughnessDelta != 2 {
		t.Fatalf("anthem delta = +%d/+%d, want +2/+2", anthem.PowerDelta, anthem.ToughnessDelta)
	}
	if !anthem.Group.Selection().OwnerNotController {
		t.Fatalf("anthem group selection = %#v, want OwnerNotController", anthem.Group.Selection())
	}

	if len(body.RuleEffects) != 1 {
		t.Fatalf("rule effects = %d, want 1", len(body.RuleEffects))
	}
	rule := body.RuleEffects[0]
	if rule.Kind != game.RuleEffectCantBeSacrificed {
		t.Fatalf("rule effect kind = %v, want RuleEffectCantBeSacrificed", rule.Kind)
	}
	if rule.AffectedController != game.ControllerYou {
		t.Fatalf("rule effect controller = %v, want ControllerYou", rule.AffectedController)
	}
	if !rule.AffectedSelection.OwnerNotController {
		t.Fatalf("rule effect selection = %#v, want OwnerNotController", rule.AffectedSelection)
	}
}
