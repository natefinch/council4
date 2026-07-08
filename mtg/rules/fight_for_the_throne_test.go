package rules

import (
	"testing"

	cardf "github.com/natefinch/council4/mtg/cards/f"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// monarchNoChoiceAgent defers every choice to its default selection. Fight for
// the Throne's delayed "you become the monarch" trigger has no targets or
// choices, so this agent is enough to drive it through the real trigger
// enumeration and resolution.
type monarchNoChoiceAgent struct{}

func (monarchNoChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (monarchNoChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

// addCreatureWithPT puts a vanilla creature with independent power and toughness
// onto controller's battlefield so a test can arrange a fight that only one of
// the two creatures survives.
func addCreatureWithPT(g *game.Game, controller game.PlayerID, power, toughness int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Custom PT Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}})
}

// castFightForThrone casts the real Fight for the Throne card as Player1,
// targeting mine ("target creature you control") and opp ("target creature an
// opponent controls"), and resolves it through the real spell-resolution path so
// the +1/+1 counter is placed, the two creatures fight, and the delayed dies
// trigger is scheduled bound to opp.
func castFightForThrone(t *testing.T, g *game.Game, engine *Engine, mine, opp *game.Permanent) {
	t.Helper()
	addImplementationSpellToStack(g, game.Player1, cardf.FightForTheThrone, []game.Target{
		game.PermanentTarget(mine.ObjectID),
		game.PermanentTarget(opp.ObjectID),
	})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Fight for the Throne was not put on the stack")
	}
	obj.TargetCounts = []int{1, 1}
	engine.resolveTopOfStack(g, &TurnLog{})
}

// newFightForThroneGame stages Player1's precombat main phase with the given
// creatures already on the battlefield, ready for Player1 to cast Fight for the
// Throne.
func newFightForThroneGame(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	return g, NewEngine(nil)
}

// TestFightForTheThroneResolutionCountersFightsAndSchedules proves the spell's
// resolution runs all three effects through the real path: the +1/+1 counter is
// placed on the controlled creature, the two creatures fight, and a delayed
// permanent-died trigger is scheduled bound to the fought opponent's creature and
// gated on the "you control your commander" intervening condition.
func TestFightForTheThroneResolutionCountersFightsAndSchedules(t *testing.T) {
	g, engine := newFightForThroneGame(t)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 5) // 5/5, becomes 6/6
	opp := addCombatCreaturePermanentWithPower(g, game.Player2, 2)  // 2/2

	castFightForThrone(t, g, engine, mine, opp)

	if got := mine.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("controlled creature +1/+1 counters = %d, want 1", got)
	}
	// The 6/6 (after the counter) and the 2/2 deal their power to each other.
	if opp.MarkedDamage != 6 {
		t.Fatalf("opponent creature marked damage = %d, want 6", opp.MarkedDamage)
	}
	if mine.MarkedDamage != 2 {
		t.Fatalf("controlled creature marked damage = %d, want 2", mine.MarkedDamage)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}
	delayed := g.DelayedTriggers[0]
	if !delayed.EventPattern.Exists || delayed.EventPattern.Val.Event != game.EventPermanentDied {
		t.Fatalf("delayed trigger pattern = %+v, want an EventPermanentDied pattern", delayed.EventPattern)
	}
	if !delayed.EventPattern.Val.DyingObjectCaptured {
		t.Fatal("delayed trigger did not capture the dying object")
	}
	if delayed.BoundDyingObjectID != opp.ObjectID {
		t.Fatalf("delayed trigger bound dying object = %v, want the fought opponent creature %v", delayed.BoundDyingObjectID, opp.ObjectID)
	}
	if !delayed.Ability.Trigger.InterveningCondition.Exists ||
		!delayed.Ability.Trigger.InterveningCondition.Val.ControllerControlsCommander {
		t.Fatalf("delayed trigger intervening condition = %+v, want ControllerControlsCommander", delayed.Ability.Trigger.InterveningCondition)
	}
}

// TestFightForTheThroneOpponentCreatureDyingWithCommanderMakesMonarch drives the
// full engine path: after the spell places the counter, resolves the fight, and
// schedules the delayed trigger, the fought opponent's creature dies to
// state-based actions this turn while Player1 controls their commander, the
// delayed trigger fires through the real enumeration, and resolving it makes
// Player1 the monarch.
func TestFightForTheThroneOpponentCreatureDyingWithCommanderMakesMonarch(t *testing.T) {
	g, engine := newFightForThroneGame(t)
	addCommanderPermanent(g, game.Player1)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 5) // 6/6 after counter
	opp := addCombatCreaturePermanentWithPower(g, game.Player2, 2)  // 2/2, takes lethal 6

	castFightForThrone(t, g, engine, mine, opp)

	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, opp.ObjectID); ok {
		t.Fatal("fought opponent creature did not die to lethal fight damage")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("delayed dies trigger did not fire when the fought creature died")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (the delayed become-monarch ability)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 did not become the monarch after the fought creature died while controlling their commander")
	}
}

// TestFightForTheThroneOpponentCreatureDyingWithoutCommanderNoMonarch proves the
// "if you control your commander" intervening condition gates the delayed
// trigger: when the fought creature dies while Player1 controls no commander, the
// trigger does not fire and Player1 does not become the monarch.
func TestFightForTheThroneOpponentCreatureDyingWithoutCommanderNoMonarch(t *testing.T) {
	g, engine := newFightForThroneGame(t)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	opp := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	castFightForThrone(t, g, engine, mine, opp)

	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, opp.ObjectID); ok {
		t.Fatal("fought opponent creature did not die to lethal fight damage")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("delayed dies trigger fired even though Player1 controlled no commander")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (no trigger)", g.Stack.Size())
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 became the monarch without controlling their commander")
	}
}

// TestFightForTheThroneDifferentCreatureDyingNoMonarch proves the delayed trigger
// is bound to the specific fought creature: when the fight leaves the opponent's
// creature alive and a different creature dies this turn, the trigger does not
// fire even though Player1 controls their commander.
func TestFightForTheThroneDifferentCreatureDyingNoMonarch(t *testing.T) {
	g, engine := newFightForThroneGame(t)
	addCommanderPermanent(g, game.Player1)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 1)  // 2/2 after counter
	opp := addCreatureWithPT(g, game.Player2, 1, 10)                 // survives the 2 damage
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 1) // 1/1 bystander

	castFightForThrone(t, g, engine, mine, opp)

	if _, ok := permanentByObjectID(g, opp.ObjectID); !ok {
		t.Fatal("the fought opponent creature should have survived the non-lethal fight")
	}
	// A different creature dies this turn.
	other.MarkedDamage = 1
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, other.ObjectID); ok {
		t.Fatal("the bystander creature did not die")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("delayed dies trigger fired when a creature other than the fought creature died")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 became the monarch from an unrelated creature's death")
	}
}

// TestFightForTheThroneCreatureDyingLaterTurnNoMonarch proves the delayed trigger
// is limited to the turn it was created: once the this-turn window ends, the
// fought creature dying on a later turn fires nothing.
func TestFightForTheThroneCreatureDyingLaterTurnNoMonarch(t *testing.T) {
	g, engine := newFightForThroneGame(t)
	addCommanderPermanent(g, game.Player1)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	opp := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	castFightForThrone(t, g, engine, mine, opp)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}

	// End-of-turn cleanup expires this-turn delayed triggers (CR 603.7b).
	expireEventDelayedTriggers(g)
	g.Turn.TurnNumber++
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after expiry = %d, want 0", len(g.DelayedTriggers))
	}

	// The fought creature (still marked with lethal fight damage) dies next turn.
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, opp.ObjectID); ok {
		t.Fatal("fought opponent creature did not die to lethal fight damage")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("expired delayed trigger fired on a later turn")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 became the monarch from an expired delayed trigger")
	}
}

// TestFightForTheThroneTokenOpponentCreatureMakesMonarch proves the delayed dies
// trigger binds correctly to a TOKEN opponent creature (CardInstanceID == 0):
// the binding resolves the fought creature's ObjectID directly from the spell's
// target, so a token that dies to the fight still makes Player1 the monarch while
// they control their commander.
func TestFightForTheThroneTokenOpponentCreatureMakesMonarch(t *testing.T) {
	g, engine := newFightForThroneGame(t)
	addCommanderPermanent(g, game.Player1)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 5) // 6/6 after counter
	opp := addCombatTokenCreaturePermanent(g, game.Player2, 2)      // 2/2 token, takes lethal 6

	castFightForThrone(t, g, engine, mine, opp)

	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, opp.ObjectID); ok {
		t.Fatal("fought opponent token did not die to lethal fight damage")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("delayed dies trigger did not fire when the fought token died")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 did not become the monarch after the fought token died while controlling their commander")
	}
}

// TestFightForTheThroneCommanderLeavingBeforeDeathNoMonarch proves the "if you
// control your commander" intervening condition is re-checked when the fought
// creature dies (CR 603.4), not when the spell resolved: a commander present at
// the fight but gone by the time the surviving fought creature dies this turn
// leaves the trigger ungated, so Player1 does not become the monarch.
func TestFightForTheThroneCommanderLeavingBeforeDeathNoMonarch(t *testing.T) {
	g, engine := newFightForThroneGame(t)
	commander := addCommanderPermanent(g, game.Player1)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 1) // 2/2 after counter
	opp := addCreatureWithPT(g, game.Player2, 1, 10)                // survives the 2 fight damage

	castFightForThrone(t, g, engine, mine, opp)
	if _, ok := permanentByObjectID(g, opp.ObjectID); !ok {
		t.Fatal("the fought opponent creature should have survived the non-lethal fight")
	}

	// The commander leaves the battlefield before the fought creature dies.
	removePermanentFromBattlefield(g, commander.ObjectID)

	// The fought creature dies later this turn.
	opp.MarkedDamage = 10
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, opp.ObjectID); ok {
		t.Fatal("the fought opponent creature did not die")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("delayed dies trigger fired even though the commander had left before the death")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 became the monarch even though they no longer controlled their commander at the death")
	}
}
