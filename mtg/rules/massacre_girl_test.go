package rules

import (
	"testing"

	cardd "github.com/natefinch/council4/mtg/cards/d"
	cardm "github.com/natefinch/council4/mtg/cards/m"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// enterMassacreGirl emits the enters-the-battlefield event for an already-placed
// Massacre Girl permanent so the real trigger enumeration fires her ETB ability.
func enterMassacreGirl(g *game.Game, mg *game.Permanent) {
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  mg.Controller,
		Player:      mg.Controller,
		PermanentID: mg.ObjectID,
		CardID:      mg.CardInstanceID,
		FromZone:    zone.None,
		ToZone:      zone.Battlefield,
	})
}

// driveDelayedDeathChain runs the state-based-action / trigger / resolution loop
// to a fixpoint, exactly as the real engine would between priority passes: deaths
// emit permanent-died events, the this-turn delayed trigger fires once per death
// (including simultaneous deaths) and resolves its body, and the cascade repeats
// until the board is stable.
func driveDelayedDeathChain(t *testing.T, engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	t.Helper()
	for range 200 {
		_, deaths := engine.applyStateBasedActionsWithDeaths(g)
		fired := engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
		resolvedAny := false
		for g.Stack.Size() > 0 {
			engine.resolveTopOfStackWithChoices(g, agents, log)
			resolvedAny = true
		}
		if len(deaths) == 0 && !fired && !resolvedAny {
			return
		}
	}
	t.Fatal("delayed death-trigger chain did not stabilize")
}

// TestMassacreGirlETBDebuffsOtherCreaturesAndSchedulesDeathTrigger proves the
// enters ability of the real Massacre Girl card runs both of its ordered effects:
// each OTHER creature gets -1/-1 until end of turn (Massacre Girl herself is
// excluded), and a this-turn delayed trigger keyed on a creature dying is
// scheduled.
func TestMassacreGirlETBDebuffsOtherCreaturesAndSchedulesDeathTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	log := TurnLog{}

	mg := addCombatPermanent(g, game.Player1, cardm.MassacreGirl())
	// Toughness >= 2 so the single ETB -1/-1 does not kill them outright, letting
	// us observe the debuff directly.
	other2 := addCreatureWithPT(g, game.Player2, 2, 2)
	other3 := addCreatureWithPT(g, game.Player1, 3, 3)

	enterMassacreGirl(g, mg)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Massacre Girl's enters trigger did not fire")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (the enters ability)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := effectivePower(g, other2); got != 1 {
		t.Fatalf("other 2/2 power after ETB = %d, want 1", got)
	}
	if got, _ := effectiveToughness(g, other2); got != 1 {
		t.Fatalf("other 2/2 toughness after ETB = %d, want 1", got)
	}
	if got := effectivePower(g, other3); got != 2 {
		t.Fatalf("other 3/3 power after ETB = %d, want 2", got)
	}
	if got, _ := effectiveToughness(g, other3); got != 2 {
		t.Fatalf("other 3/3 toughness after ETB = %d, want 2", got)
	}
	// Massacre Girl is excluded from her own -1/-1.
	if got := effectivePower(g, mg); got != 4 {
		t.Fatalf("Massacre Girl power after ETB = %d, want 4 (excluded from her own debuff)", got)
	}
	if got, _ := effectiveToughness(g, mg); got != 4 {
		t.Fatalf("Massacre Girl toughness after ETB = %d, want 4 (excluded from her own debuff)", got)
	}

	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}
	delayed := g.DelayedTriggers[0]
	if !delayed.EventPattern.Exists || delayed.EventPattern.Val.Event != game.EventPermanentDied {
		t.Fatalf("delayed trigger pattern = %+v, want an EventPermanentDied pattern", delayed.EventPattern)
	}
}

// TestMassacreGirlDeathTriggerFiresOnEachDeath proves the this-turn delayed
// trigger fires when a creature dies and applies another mass -1/-1: after the
// ETB debuff, killing the weakened 2/2 (now 1/1) via state-based actions fires
// the delayed trigger, and resolving it drops every other creature by a further
// -1/-1.
func TestMassacreGirlDeathTriggerFiresOnEachDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	log := TurnLog{}

	mg := addCombatPermanent(g, game.Player1, cardm.MassacreGirl())
	dying := addCreatureWithPT(g, game.Player2, 2, 2) // 2/2 -> 1/1 after ETB
	bystander := addCreatureWithPT(g, game.Player2, 4, 4)

	enterMassacreGirl(g, mg)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Massacre Girl's enters trigger did not fire")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	// The bystander is at 3/3 after the ETB; the dying 2/2 is at 1/1. Reduce the
	// dying creature to 0 toughness with an extra point of marked damage so it dies
	// to state-based actions, standing in for combat/other death this turn.
	dying.MarkedDamage = 1
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, dying.ObjectID); ok {
		t.Fatal("the weakened creature is still on the battlefield")
	}

	// The creature death this turn fires the delayed trigger.
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("the creature death did not fire Massacre Girl's delayed -1/-1 trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	// The bystander has now taken two -1/-1s: the ETB and the death trigger.
	if got, _ := effectiveToughness(g, bystander); got != 2 {
		t.Fatalf("bystander toughness after death trigger = %d, want 2 (4 minus two -1/-1)", got)
	}
	if got := effectivePower(g, bystander); got != 2 {
		t.Fatalf("bystander power after death trigger = %d, want 2 (4 minus two -1/-1)", got)
	}
}

// TestMassacreGirlChainWipesEveryOtherCreature proves the signature Massacre Girl
// board wipe: the ETB -1/-1 kills the smallest creature, whose death fires the
// delayed trigger for another -1/-1, which kills the next creature, and so on
// until every other creature (across every player, including Massacre Girl's own
// controller) is dead while Massacre Girl herself survives.
func TestMassacreGirlChainWipesEveryOtherCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	log := TurnLog{}

	mg := addCombatPermanent(g, game.Player1, cardm.MassacreGirl())
	victims := []*game.Permanent{
		addCreatureWithPT(g, game.Player2, 1, 1),
		addCreatureWithPT(g, game.Player2, 2, 2),
		addCreatureWithPT(g, game.Player2, 3, 3),
		addCreatureWithPT(g, game.Player1, 3, 3), // Massacre Girl's own creature dies too.
	}

	enterMassacreGirl(g, mg)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Massacre Girl's enters trigger did not fire")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	driveDelayedDeathChain(t, engine, g, agents, &log)

	for i, victim := range victims {
		if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
			t.Fatalf("victim %d survived the Massacre Girl chain, want dead", i)
		}
	}
	if _, ok := permanentByObjectID(g, mg.ObjectID); !ok {
		t.Fatal("Massacre Girl died to her own chain, want alive (she is excluded from every -1/-1)")
	}
	if got, _ := effectiveToughness(g, mg); got != 4 {
		t.Fatalf("Massacre Girl toughness after the wipe = %d, want 4 (never debuffed)", got)
	}
}

// TestMassacreGirlWipesGoWideBoardIncludingLargerCreature is the regression test
// for the simultaneous-death fan-out fix. Massacre Girl's ETB -1/-1 drops four
// 1/1s to 0/0 so they die together in one state-based-action batch. Each of those
// four simultaneous deaths must fire the this-turn delayed trigger, stacking four
// more board-wide -1/-1s that, together with the ETB debuff, reduce a 5/5 to 0/0
// and kill it. Before the fix the batch fired the trigger only once, leaving the
// 5/5 alive at 3/3, so this test would fail.
func TestMassacreGirlWipesGoWideBoardIncludingLargerCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	log := TurnLog{}

	mg := addCombatPermanent(g, game.Player1, cardm.MassacreGirl())
	small := []*game.Permanent{
		addCreatureWithPT(g, game.Player2, 1, 1),
		addCreatureWithPT(g, game.Player2, 1, 1),
		addCreatureWithPT(g, game.Player2, 1, 1),
		addCreatureWithPT(g, game.Player2, 1, 1),
	}
	big := addCreatureWithPT(g, game.Player2, 5, 5)

	enterMassacreGirl(g, mg)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Massacre Girl's enters trigger did not fire")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	driveDelayedDeathChain(t, engine, g, agents, &log)

	for i, victim := range small {
		if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
			t.Fatalf("small creature %d survived the go-wide chain, want dead", i)
		}
	}
	if _, ok := permanentByObjectID(g, big.ObjectID); ok {
		t.Fatal("the 5/5 survived: the four simultaneous 1/1 deaths must each fire the delayed -1/-1 to accumulate lethal debuff")
	}
	if _, ok := permanentByObjectID(g, mg.ObjectID); !ok {
		t.Fatal("Massacre Girl died to her own chain, want alive (she is excluded from every -1/-1)")
	}
}

// TestDeathFrenzyGainsLifePerSimultaneousDeath proves the fan-out fix generalizes
// beyond Massacre Girl: Death Frenzy's -2/-2 kills four 2/2 creatures at once, and
// its "whenever a creature dies this turn, you gain 1 life" delayed trigger must
// fire once per creature, gaining 4 life. Before the fix the single batch fired
// the trigger once and gained only 1 life.
func TestDeathFrenzyGainsLifePerSimultaneousDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	log := TurnLog{}

	// Death Frenzy's -2/-2 hits every creature; four 2/2s all drop to 0/0 and die
	// together, spread across players to show the trigger counts every death.
	victims := []*game.Permanent{
		addCreatureWithPT(g, game.Player2, 2, 2),
		addCreatureWithPT(g, game.Player2, 2, 2),
		addCreatureWithPT(g, game.Player1, 2, 2),
		addCreatureWithPT(g, game.Player3, 2, 2),
	}
	g.Players[game.Player1].Life = 20

	addImplementationSpellToStack(g, game.Player1, cardd.DeathFrenzy(), nil)
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	driveDelayedDeathChain(t, engine, g, agents, &log)

	for i, victim := range victims {
		if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
			t.Fatalf("victim %d survived Death Frenzy's -2/-2, want dead", i)
		}
	}
	if got := g.Players[game.Player1].Life; got != 24 {
		t.Fatalf("Death Frenzy controller life = %d, want 24 (20 + 1 life for each of 4 simultaneous deaths)", got)
	}
}
