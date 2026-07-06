package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// championsOfMinasTirithOracleText is the printed Champions of Minas Tirith rules
// text. The first ability lowers to BecomeMonarch; the second lowers to the
// PlayerMayPayGenericOrRule punisher gated by the "if you're the monarch"
// intervening condition.
const championsOfMinasTirithOracleText = "When this creature enters, you become the monarch.\n" +
	"At the beginning of combat on each opponent's turn, if you're the monarch, " +
	"that opponent may pay {X}, where X is the number of cards in their hand. " +
	"If they don't, they can't attack you this combat."

func TestLowerChampionsOfMinasTirithPunisher(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Champions of Minas Tirith",
		Layout:     "normal",
		ManaCost:   "{5}{W}",
		TypeLine:   "Creature — Human Soldier",
		OracleText: championsOfMinasTirithOracleText,
	})

	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want monarch ETB and combat punisher", len(face.TriggeredAbilities))
	}
	punisher := face.TriggeredAbilities[1]
	trigger := punisher.Trigger
	if trigger.Type != game.TriggerAt ||
		trigger.Pattern.Event != game.EventBeginningOfStep ||
		trigger.Pattern.Step != game.StepBeginningOfCombat ||
		trigger.Pattern.Controller != game.TriggerControllerOpponent {
		t.Fatalf("trigger = %#v, want each opponent's beginning of combat", trigger)
	}
	if !trigger.InterveningCondition.Exists || !trigger.InterveningCondition.Val.ControllerIsMonarch {
		t.Fatalf("intervening condition = %#v, want ControllerIsMonarch", trigger.InterveningCondition)
	}

	if len(punisher.Content.Modes) != 1 {
		t.Fatalf("punisher modes = %d, want one", len(punisher.Content.Modes))
	}
	sequence := punisher.Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want a single pay-or-rule instruction", len(sequence))
	}
	prim, ok := sequence[0].Primitive.(game.PlayerMayPayGenericOrRule)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want PlayerMayPayGenericOrRule", sequence[0].Primitive)
	}
	if prim.Player.Kind() != game.PlayerReferenceEventPlayer {
		t.Fatalf("payer = %#v, want the triggering event player", prim.Player)
	}
	dynamic := prim.Amount.DynamicAmount()
	if !dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountCountCardsInZone ||
		dynamic.Val.CardZone != zone.Hand ||
		dynamic.Val.Player == nil ||
		dynamic.Val.Player.Kind() != game.PlayerReferenceEventPlayer {
		t.Fatalf("amount = %#v, want the event player's hand size", prim.Amount)
	}
	if prim.Duration != game.DurationUntilEndOfCombat {
		t.Fatalf("duration = %#v, want until end of combat", prim.Duration)
	}
	if len(prim.RuleEffects) != 1 {
		t.Fatalf("rule effects = %d, want one can't-attack effect", len(prim.RuleEffects))
	}
	effect := prim.RuleEffects[0]
	if effect.Kind != game.RuleEffectCantAttack ||
		effect.AffectedPlayerRef.Kind() != game.PlayerReferenceEventPlayer ||
		effect.DefendingPlayer != game.PlayerYou ||
		!effect.DefendingPlayerDirectOnly {
		t.Fatalf("rule effect = %#v, want the event player's creatures can't attack you directly", effect)
	}
}
