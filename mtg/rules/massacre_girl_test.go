package rules

import (
	"slices"
	"testing"

	cardd "github.com/natefinch/council4/mtg/cards/d"
	cardm "github.com/natefinch/council4/mtg/cards/m"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
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

// Source review (#3502): Oracle text from the pinned corpus for Massacre Girl
// (59957c45-4672-4650-80b1-70e4a60b429e, DSC 146): "Menace / When Massacre
// Girl enters, each other creature gets -1/-1 until end of turn. Whenever a
// creature dies this turn, each creature other than Massacre Girl gets -1/-1
// until end of turn." https://api.scryfall.com/cards/59957c45-4672-4650-80b1-70e4a60b429e/rulings
// returned five 2019-05-03 rulings: both abilities affect only creatures on the
// battlefield as they resolve; once the enters ability has triggered it doesn't
// matter whether Massacre Girl remains on the battlefield; deaths before the
// delayed trigger is created don't trigger it; 0-toughness creatures die after
// the delayed trigger exists; simultaneous deaths trigger it that many times.
// CR 2026-09-25 (official TXT, MagicCompRules 20260925): 201.5 (a card's own
// name means that particular object), 400.7 (a zone change makes a new
// object), 603.7a/603.7c/603.7e (a delayed trigger created by a triggered
// ability shares its source and acts on the particular object it refers to,
// not a later object for the same card) and 611.2c (the affected set is fixed
// when the effect begins). The delayed trigger therefore binds the Massacre
// Girl object whose ability created it, keeps that binding observable after it
// leaves, still excludes it while present, and excludes no later object.

// massacreGirlScheduled puts the default-registry Massacre Girl onto the
// battlefield with the given other creatures and resolves her enters ability,
// scheduling her this-turn delayed death trigger.
func massacreGirlScheduled(t *testing.T, others ...*game.Permanent) (*game.Game, *Engine, *game.Permanent, [game.NumPlayers]PlayerAgent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	mg := addCombatPermanent(g, game.Player1, registryCard(t, "Massacre Girl"))
	enterMassacreGirl(g, mg)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Massacre Girl's enters trigger did not fire")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want Massacre Girl's death trigger", len(g.DelayedTriggers))
	}
	return g, engine, mg, agents
}

func assertPT(t *testing.T, g *game.Game, permanent *game.Permanent, want int, what string) {
	t.Helper()
	power := effectivePower(g, permanent)
	toughness, _ := effectiveToughness(g, permanent)
	if power != want || toughness != want {
		t.Fatalf("%s = %d/%d, want %d/%d", what, power, toughness, want, want)
	}
}

// While Massacre Girl remains, her delayed trigger is bound to her pending and
// on the stack, and resolving it still excludes her.
func TestMassacreGirlDelayedTriggerBindsAndExcludesPresentSource(t *testing.T) {
	g, engine, mg, agents := massacreGirlScheduled(t)
	victim := addCreatureWithPT(g, game.Player2, 1, 1)
	bystander := addCreatureWithPT(g, game.Player2, 4, 4)

	obs := observe(g, game.Player2)
	if delayedViewBoundTo(obs.DelayedTriggers(), mg.ObjectID) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want one bound to Massacre Girl %d", obs.DelayedTriggers(), mg.ObjectID)
	}
	if delayedViewBoundTo(findPermanentView(t, obs, mg.ObjectID).PendingDelayedTriggers, mg.ObjectID) != 1 {
		t.Fatal("Massacre Girl's view does not list her bound delayed trigger")
	}

	victim.MarkedDamage = 1
	engine.applyStateBasedActions(g)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the creature death did not fire Massacre Girl's delayed trigger")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || !slices.Contains(stack[0].BoundObjectIDs, mg.ObjectID) || stack[0].SourceObjectID != mg.ObjectID {
		t.Fatalf("stack = %+v, want the delayed trigger sourced from and bound to Massacre Girl %d", stack, mg.ObjectID)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertPT(t, g, bystander, 3, "bystander after the delayed -1/-1")
	assertPT(t, g, mg, 4, "Massacre Girl after her delayed -1/-1")
}

// After Massacre Girl leaves the battlefield, her pending delayed trigger and
// the stack object it becomes still name the original object (CR 603.7c), and
// it keeps working.
func TestMassacreGirlDelayedTriggerKeepsBindingAfterDeparture(t *testing.T) {
	g, engine, mg, agents := massacreGirlScheduled(t)
	victim := addCreatureWithPT(g, game.Player2, 1, 1)
	bystander := addCreatureWithPT(g, game.Player2, 4, 4)
	oldID := mg.ObjectID
	if !movePermanentToZone(g, mg, zone.Exile) {
		t.Fatal("could not exile Massacre Girl")
	}

	views := observe(g, game.Player2).DelayedTriggers()
	if len(views) != 1 || !slices.Equal(views[0].BoundObjectIDs, []id.ID{oldID}) || views[0].SourceObjectID != 0 || views[0].SourceName != "Massacre Girl" {
		t.Fatalf("pending delayed triggers = %+v, want Massacre Girl's trigger still bound to %d with no battlefield source", views, oldID)
	}

	victim.MarkedDamage = 1
	engine.applyStateBasedActions(g)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the creature death did not fire Massacre Girl's delayed trigger after she left")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 1 || !slices.Equal(stack[0].BoundObjectIDs, []id.ID{oldID}) || stack[0].SourceObjectID != 0 {
		t.Fatalf("stack = %+v, want the fired trigger still bound to departed Massacre Girl %d", stack, oldID)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertPT(t, g, bystander, 3, "bystander after the delayed -1/-1")
}

// A Massacre Girl that leaves and returns is a new object (CR 400.7). The
// original's delayed trigger does not exclude it, while the new object's own
// enters ability creates a separate delayed trigger that does.
func TestMassacreGirlDelayedTriggerDoesNotExcludeLaterObject(t *testing.T) {
	g, engine, mg, agents := massacreGirlScheduled(t)
	victim := addCreatureWithPT(g, game.Player2, 1, 1)
	bystander := addCreatureWithPT(g, game.Player2, 5, 5)
	oldID := mg.ObjectID
	if !movePermanentToZone(g, mg, zone.Exile) {
		t.Fatal("could not exile Massacre Girl")
	}
	card, ok := g.GetCardInstance(mg.CardInstanceID)
	if !ok {
		t.Fatal("Massacre Girl card instance not found")
	}
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok || returned.ObjectID == oldID {
		t.Fatal("Massacre Girl did not return as a new object")
	}
	obs := observe(g, game.Player2)
	if delayedViewBoundTo(obs.DelayedTriggers(), oldID) != 1 {
		t.Fatalf("pending delayed triggers = %+v, want the original's trigger bound to %d", obs.DelayedTriggers(), oldID)
	}
	if pending := findPermanentView(t, obs, returned.ObjectID).PendingDelayedTriggers; len(pending) != 0 {
		t.Fatalf("returned Massacre Girl pending = %+v, want none from the original object", pending)
	}

	// The returned object's own enters ability: the 1/1 victim drops to 0/0 and
	// the 5/5 bystander to 4/4, and a second delayed trigger binds the new object.
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the returned Massacre Girl's enters trigger did not fire")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if delayedViewBoundTo(observe(g, game.Player2).DelayedTriggers(), returned.ObjectID) != 1 {
		t.Fatal("the returned Massacre Girl did not schedule a delayed trigger bound to itself")
	}

	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("the 0/0 victim did not die")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("the death did not fire the delayed triggers")
	}
	stack := observe(g, game.Player2).Stack()
	if len(stack) != 2 || delayedStackBoundTo(stack, oldID) != 1 || delayedStackBoundTo(stack, returned.ObjectID) != 1 {
		t.Fatalf("stack = %+v, want one trigger bound to the original %d and one to the returned %d", stack, oldID, returned.ObjectID)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	// Only the original's trigger affects the returned object; both hit the bystander.
	assertPT(t, g, returned, 3, "returned Massacre Girl")
	assertPT(t, g, bystander, 2, "bystander")
}

func delayedStackBoundTo(stack []StackObjectView, objectID id.ID) int {
	count := 0
	for _, view := range stack {
		if slices.Contains(view.BoundObjectIDs, objectID) {
			count++
		}
	}
	return count
}
