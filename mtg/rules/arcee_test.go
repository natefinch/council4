package rules

import (
	"testing"

	cardsa "github.com/natefinch/council4/mtg/cards/a"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// newArceeFront puts the real Arcee, Sharpshooter card onto the controller's
// battlefield as its front face so its "{1}, Remove one or more +1/+1 counters
// from Arcee: It deals that much damage to target creature. Convert Arcee."
// activated ability runs through the real activation, cost-payment, and
// resolution paths.
func newArceeFront(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cardsa.ArceeSharpshooter())
	permanent.Face = game.FaceFront
	return permanent
}

// newArceeBack puts the real Arcee onto the controller's battlefield as its back
// face (Arcee, Acrobatic Coupe) so its "Whenever you cast a spell that targets
// one or more creatures or Vehicles you control, put that many +1/+1 counters on
// Arcee. Convert Arcee." trigger runs through the real detection and resolution
// paths.
func newArceeBack(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cardsa.ArceeSharpshooter())
	permanent.Face = game.FaceBack
	permanent.Transformed = true
	return permanent
}

// TestArceeSharpshooterVariableCounterRemovalScalesDamageAndConverts proves the
// variable +1/+1-counter-removal activation cost (issue #2847's high-leverage
// mechanic): the player chooses how many counters to remove (one or more), that
// chosen count is removed as an additional cost, and the damage the ability
// deals equals exactly that many. It then converts Arcee.
func TestArceeSharpshooterVariableCounterRemovalScalesDamageAndConverts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	arcee := newArceeFront(g, game.Player1)
	arcee.Counters.Add(counter.PlusOnePlusOne, 3)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	g.Turn.PriorityPlayer = game.Player1

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}

	// "one or more" forbids removing zero: X=0 must never be a legal activation.
	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(arcee.ObjectID, 0, targets, 0)) {
		t.Fatal("X=0 activation was legal; 'one or more' requires removing at least one counter")
	}
	// The chosen count cannot exceed the counters present (only 3 counters).
	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(arcee.ObjectID, 0, targets, 4)) {
		t.Fatal("X=4 activation was legal with only 3 counters present")
	}

	act := action.ActivateAbility(arcee.ObjectID, 0, targets, 2)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("removing two counters (X=2) was not a legal activation")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(activate Arcee) = false, want true")
	}
	// The two counters are removed as the additional cost when the ability goes
	// on the stack, before it resolves.
	if got := arcee.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("Arcee +1/+1 counters after paying cost = %d, want 1 (3 - 2 removed)", got)
	}

	engine.resolveTopOfStack(g, nil)

	if target.MarkedDamage != 2 {
		t.Fatalf("target marked damage = %d, want 2 (equal to counters removed)", target.MarkedDamage)
	}
	if arcee.Face != game.FaceBack || !arcee.Transformed {
		t.Fatalf("Arcee face/transformed = %v/%v, want back/true (Convert Arcee)", arcee.Face, arcee.Transformed)
	}
}

// TestArceeAcrobaticCoupeCountsSpellTargetsPutsCountersAndConverts proves the
// back-face cast trigger: casting a spell that targets one or more of your
// creatures or Vehicles puts exactly that many +1/+1 counters on Arcee (counting
// only your creatures/Vehicles among the spell's targets, not the opponent's)
// and converts Arcee.
func TestArceeAcrobaticCoupeCountsSpellTargetsPutsCountersAndConverts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	arcee := newArceeBack(g, game.Player1)

	mine1 := addCreaturePermanent(g, game.Player1)
	mine2 := addCreaturePermanent(g, game.Player1)
	opponent := addCreaturePermanent(g, game.Player2)

	// A spell you cast targeting two of your creatures and one opponent creature:
	// only the two you control count toward "that many".
	castSpellTargeting(g, game.Player1,
		game.PermanentTarget(mine1.ObjectID),
		game.PermanentTarget(mine2.ObjectID),
		game.PermanentTarget(opponent.ObjectID))
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger did not fire on a spell targeting your creatures")
	}

	engine.resolveTopOfStack(g, nil)

	if got := arcee.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("Arcee +1/+1 counters = %d, want 2 (only your two targeted creatures count)", got)
	}
	if arcee.Face != game.FaceFront || arcee.Transformed {
		t.Fatalf("Arcee face/transformed = %v/%v, want front/false (Convert Arcee)", arcee.Face, arcee.Transformed)
	}
}
