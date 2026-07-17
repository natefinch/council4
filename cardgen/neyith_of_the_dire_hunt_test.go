package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
)

const neyithOracleText = "Whenever one or more creatures you control fight or become blocked, draw a card.\n" +
	"At the beginning of combat on your turn, you may pay {2}{R/G}. If you do, double target creature's power until end of turn. That creature must be blocked this combat if able. ({R/G} can be paid with either {R} or {G}.)"

func TestLowerNeyithOfTheDireHunt(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Neyith of the Dire Hunt",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Warrior",
		OracleText: neyithOracleText,
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %#v, want draw trigger and combat trigger", face.TriggeredAbilities)
	}

	drawPattern := face.TriggeredAbilities[0].Trigger.Pattern
	if drawPattern.Event != game.EventFight ||
		drawPattern.UnionEvent != game.EventAttackerBecameBlocked ||
		drawPattern.Controller != game.TriggerControllerYou ||
		!drawPattern.OneOrMore {
		t.Fatalf("draw trigger pattern = %#v", drawPattern)
	}

	combat := face.TriggeredAbilities[1]
	if combat.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		combat.Trigger.Pattern.Step != game.StepBeginningOfCombat ||
		combat.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("combat trigger pattern = %#v", combat.Trigger.Pattern)
	}
	if len(combat.Content.Modes) != 1 {
		t.Fatalf("combat modes = %#v, want one", combat.Content.Modes)
	}
	mode := combat.Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 3 {
		t.Fatalf("combat mode = %#v, want one target and pay/double/must-block", mode)
	}
	pay, ok := mode.Sequence[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("instruction[0] = %T, want game.Pay", mode.Sequence[0].Primitive)
	}
	wantCost := cost.Mana{cost.O(2), cost.HybridMana(mana.R, mana.G)}
	if !pay.Payment.ManaCost.Exists || !slices.Equal(pay.Payment.ManaCost.Val, wantCost) {
		t.Fatalf("payment cost = %#v, want %v", pay.Payment.ManaCost, wantCost)
	}
	if mode.Sequence[0].PublishResult != controllerPaidResultKey {
		t.Fatalf("payment result = %q, want %q", mode.Sequence[0].PublishResult, controllerPaidResultKey)
	}
	double, ok := mode.Sequence[1].Primitive.(game.ModifyPT)
	if !ok || double.Duration != game.DurationUntilEndOfTurn ||
		double.Object != game.TargetPermanentReference(0) {
		t.Fatalf("instruction[1] = %#v, want target power doubling until end of turn", mode.Sequence[1])
	}
	power := double.PowerDelta.DynamicAmount()
	if !power.Exists ||
		power.Val.Kind != game.DynamicAmountObjectPower ||
		power.Val.Object != game.TargetPermanentReference(0) {
		t.Fatalf("power delta = %#v, want target power at resolution", double.PowerDelta)
	}
	mustBlock, ok := mode.Sequence[2].Primitive.(game.ApplyRule)
	if !ok || mustBlock.Duration != game.DurationUntilEndOfCombat ||
		!mustBlock.Object.Exists || mustBlock.Object.Val != game.TargetPermanentReference(0) ||
		len(mustBlock.RuleEffects) != 1 ||
		mustBlock.RuleEffects[0].Kind != game.RuleEffectMustBeBlocked {
		t.Fatalf("instruction[2] = %#v, want target must be blocked this combat", mode.Sequence[2])
	}
	for i := 1; i < len(mode.Sequence); i++ {
		gate := mode.Sequence[i].ResultGate
		if !gate.Exists || gate.Val.Key != controllerPaidResultKey || gate.Val.Succeeded != game.TriTrue {
			t.Fatalf("instruction[%d] gate = %#v, want successful payment gate", i, gate)
		}
	}
}
