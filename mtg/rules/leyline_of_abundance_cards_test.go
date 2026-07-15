package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLeylineOfAbundanceAddsGreenWhenControllerTapsCreatureForMana proves the
// "Whenever you tap a creature for mana, add an additional {G}." trigger fires
// and adds one green mana when the controller taps their own creature for mana.
func TestLeylineOfAbundanceAddsGreenWhenControllerTapsCreatureForMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfAbundance())
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	setPermanentTappedForMana(g, creature)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("tap-for-mana trigger did not fire for the controller's creature")
	}
	engine.resolveTopOfStack(g, nil)

	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana added = %d, want 1", got)
	}
}

// TestLeylineOfAbundanceDoesNotTriggerForOpponentCreature proves the trigger is
// controller-scoped: an opponent tapping their creature for mana does not fire
// the controller's Leyline of Abundance.
func TestLeylineOfAbundanceDoesNotTriggerForOpponentCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfAbundance())
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	setPermanentTappedForMana(g, opponentCreature)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("tap-for-mana trigger fired for an opponent-controlled creature")
	}
}

// TestLeylineOfAbundanceDoesNotTriggerForNoncreatureMana proves the trigger's
// creature requirement: tapping a land for mana does not fire it.
func TestLeylineOfAbundanceDoesNotTriggerForNoncreatureMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfAbundance())
	land := addBasicLandPermanent(g, game.Player1, types.Forest)

	setPermanentTappedForMana(g, land)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("tap-for-mana trigger fired for a noncreature (land) tapped for mana")
	}
}

// TestLeylineOfAbundanceActivatedCountersEachCreatureYouControl proves the
// "{6}{G}{G}: Put a +1/+1 counter on each creature you control." ability pays
// its mana, puts a counter on each of the controller's creatures, and leaves
// opponents' creatures untouched.
func TestLeylineOfAbundanceActivatedCountersEachCreatureYouControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	abundance := addCombatPermanent(g, game.Player1, cards.LeylineOfAbundance())
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	for range 8 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.ActivateAbility(abundance.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Leyline of Abundance activated ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(activate) = false, want true")
	}
	engine.resolveTopOfStack(g, nil)

	if got := first.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("first creature +1/+1 counters = %d, want 1", got)
	}
	if got := second.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("second creature +1/+1 counters = %d, want 1", got)
	}
	if got := opponentCreature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("opponent creature +1/+1 counters = %d, want 0", got)
	}
}
