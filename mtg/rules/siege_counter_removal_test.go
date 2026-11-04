package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3649 (sub-issue of #3596), a Siege's intrinsic defeat
// trigger when an effect, cost or counter move removes its last defense
// counter. Official CR effective 2026-09-25 (verified local copy), read 122.1g,
// 122.5, 122.8, 310.12b, 601.2h, 603.2c, 704.5v. 122.1g/704.5v: a battle with 0
// defense isn't put into its owner's graveyard while it is the source of an
// ability that has triggered but not yet left the stack; 122.5: moving a counter
// removes it from the first object and puts it onto the second; 122.8: when the
// first object has left the battlefield, no counter is removed; 310.12b: each
// Siege has "When the last defense counter is removed from this permanent,
// exile it, then you may cast it transformed without paying its mana cost.".
// Cards (pinned corpus): Vampire Hexmage (oracle
// fcb2a95b-cef0-4b29-8b2b-538f380aff2b, {B}{B} 2/1, "First strike\nSacrifice
// this creature: Remove all counters from target permanent."); Invasion of
// Segovia // Caetus, Sea Tyrant of Segovia (oracle
// 7b3c0d89-55ed-43d7-a00a-7b060857ec4e, {2}{U} Battle — Siege, defense 4; back
// Legendary Creature — Serpent 3/3); Etched Host Doombringer (oracle
// 85c799e9-74d5-459e-b168-b95ed78cd4e9, "... Choose target battle. If an
// opponent protects it, remove three defense counters from it. ..."; no
// rulings), whose resolving removal is the RemoveCounter route proven here.
// Scryfall rulings: Vampire Hexmage 2021-03-19 (any permanent can be targeted,
// only on the battlefield); Invasion of Segovia 2023-04-14 (the ability triggers
// when the last defense counter is removed; a Siege that never had defense
// counters can't have its last one removed).

// defeatedOrWaitingSiege applies state-based actions, proves the 0-defense
// Siege is still on the battlefield while its triggered ability waits to be put
// on the stack (CR 704.5v), then puts triggered abilities on the stack.
func defeatedOrWaitingSiege(t *testing.T, g *game.Game, engine *Engine, sieges ...*game.Permanent) {
	t.Helper()
	engine.applyStateBasedActions(g)
	for _, siege := range sieges {
		if _, ok := permanentByObjectID(g, siege.ObjectID); !ok {
			t.Fatal("Siege was put into the graveyard while its defeat ability was waiting to be put on the stack")
		}
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func countersRemovedEvents(g *game.Game) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventCountersRemoved {
			events = append(events, event)
		}
	}
	return events
}

// TestHexmageRemovingAllCountersDefeatsSegovia is acceptance scenario 1:
// Player1 sacrifices registered Vampire Hexmage targeting Player2's registered
// Invasion of Segovia, which also carries a charge counter.
func TestHexmageRemovingAllCountersDefeatsSegovia(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	engine := choicesProgramEngine(t)
	hexmageDef := registryCardDef(t, "Vampire Hexmage")
	hexmage := addCombatPermanent(g, game.Player1, hexmageDef)
	siege := addCombatPermanent(g, game.Player2, registryCardDef(t, "Invasion of Segovia"))
	siege.Counters.Add(counter.Defense, 2)
	siege.Counters.Add(counter.Charge, 1)
	siege.BattleProtector = opt.Val(game.Player1)
	cardID := siege.CardInstanceID

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(hexmage.ObjectID,
		hexmageDef.ActivatedAbilityIndex(0), []game.Target{game.PermanentTarget(siege.ObjectID)}, 0)) {
		t.Fatal("Vampire Hexmage's ability could not target Invasion of Segovia")
	}
	if !g.Players[game.Player1].Graveyard.Contains(hexmage.CardInstanceID) {
		t.Fatal("Vampire Hexmage was not sacrificed to pay its cost")
	}
	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !siege.Counters.IsEmpty() {
		t.Fatalf("Siege counters = %v, want none", siege.Counters.All())
	}
	events := countersRemovedEvents(g)
	if len(events) != 2 || events[0].SimultaneousID == 0 || events[0].SimultaneousID != events[1].SimultaneousID {
		t.Fatalf("counters removed events = %+v, want one simultaneous event per kind", events)
	}
	for _, event := range events {
		if event.PermanentID != siege.ObjectID || event.Controller != game.Player1 || event.Amount != event.PreviousCounterAmount {
			t.Fatalf("counters removed event = %+v, want all of a kind removed by Player1's Hexmage", event)
		}
	}

	defeatedOrWaitingSiege(t, g, engine, siege)
	siegeDefeatAbilityOnTop(t, g, siege)
	if siegeDefeatAbilities(g) != 1 {
		t.Fatalf("stack = %+v, want exactly one defeat ability", g.Stack.Objects())
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, siege.ObjectID); !ok {
		t.Fatal("Siege was put into the graveyard while its defeat ability was on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	spell, ok := g.Stack.Peek()
	if !ok || spell.Kind != game.StackSpell || spell.SourceID != cardID || spell.Face != game.FaceBack ||
		spell.Controller != game.Player2 || spell.ManaSpentToCast != 0 {
		t.Fatalf("stack top = %+v, want Caetus cast transformed for free by Player2", spell)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	caetus := permanentForCard(g, cardID)
	if caetus == nil || caetus.Face != game.FaceBack || !caetus.Transformed || caetus.Controller != game.Player2 {
		t.Fatalf("Caetus = %+v, want it on the battlefield transformed under Player2's control", caetus)
	}
}

// resolvingDefenseRemoval has Player2 resolve "remove amount defense counters
// from target battle" against Player1's registered Invasion of Segovia with the
// given defense, the route of Etched Host Doombringer's second mode.
func resolvingDefenseRemoval(t *testing.T, defense, amount int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := mainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	siege := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	siege.Counters.Add(counter.Defense, defense)
	siege.BattleProtector = opt.Val(game.Player2)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackTriggeredAbility,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(siege.ObjectID)},
	}
	resolveInstruction(engine, g, obj, game.RemoveCounter{
		Amount:      game.Fixed(amount),
		Object:      game.TargetPermanentReference(0),
		CounterKind: counter.Defense,
	}, &TurnLog{})
	return g, engine, siege
}

// TestResolvingDefenseRemovalOfLastCountersTriggersSiege is acceptance
// scenarios 2 and 3: removing exactly the remaining defense triggers the Siege,
// and state-based actions leave the 0-defense battle while it waits.
func TestResolvingDefenseRemovalOfLastCountersTriggersSiege(t *testing.T) {
	g, engine, siege := resolvingDefenseRemoval(t, 3, 3)
	removed := lastCountersRemovedEvent(t, g)
	if removed.PermanentID != siege.ObjectID || removed.CounterKind != counter.Defense ||
		removed.PreviousCounterAmount != 3 || removed.Amount != 3 || removed.Controller != game.Player2 ||
		removed.SimultaneousID != 0 {
		t.Fatalf("counters removed event = %+v, want 3 of 3 defense counters removed by Player2", removed)
	}
	defeatedOrWaitingSiege(t, g, engine, siege)
	siegeDefeatAbilityOnTop(t, g, siege)
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, siege.ObjectID); !ok {
		t.Fatal("Siege was put into the graveyard while its defeat ability was on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if spell, ok := g.Stack.Peek(); !ok || spell.SourceID != siege.CardInstanceID || spell.Face != game.FaceBack || spell.Controller != game.Player1 {
		t.Fatalf("stack top = %+v, want Caetus cast transformed by the Siege's controller", spell)
	}
}

// TestResolvingDefenseRemovalLeavingDefenseDoesNotTriggerSiege is acceptance
// scenario 2's other half: removing fewer than the remaining defense counters
// publishes the removal but doesn't trigger the Siege.
func TestResolvingDefenseRemovalLeavingDefenseDoesNotTriggerSiege(t *testing.T) {
	g, engine, siege := resolvingDefenseRemoval(t, 4, 3)
	if defense := siege.Counters.Get(counter.Defense); defense != 1 {
		t.Fatalf("Siege defense = %d, want 1", defense)
	}
	if removed := lastCountersRemovedEvent(t, g); removed.PreviousCounterAmount != 4 || removed.Amount != 3 {
		t.Fatalf("counters removed event = %+v, want 3 of 4", removed)
	}
	engine.applyStateBasedActions(g)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if siegeDefeatAbilities(g) != 0 {
		t.Fatalf("stack = %+v, want no defeat ability", g.Stack.Objects())
	}
	if _, ok := permanentByObjectID(g, siege.ObjectID); !ok {
		t.Fatal("Siege with defense left the battlefield")
	}
}

// TestGroupDefenseRemovalTriggersEachDefeatedSiege proves one group removal
// that reduces several Sieges to 0 defense triggers each of them, while a Siege
// left with defense and a non-Siege battle don't trigger.
func TestGroupDefenseRemovalTriggersEachDefeatedSiege(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	segovia := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	kylem := addCombatPermanent(g, game.Player2, registryCardDef(t, "Invasion of Kylem"))
	survivor := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	plain := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Plain Battle", Types: []types.Card{types.Battle}, Defense: opt.Val(2),
	}})
	segovia.Counters.Add(counter.Defense, 2)
	kylem.Counters.Add(counter.Defense, 1)
	survivor.Counters.Add(counter.Defense, 3)
	plain.Counters.Add(counter.Defense, 2)
	segovia.BattleProtector = opt.Val(game.Player2)
	kylem.BattleProtector = opt.Val(game.Player1)
	survivor.BattleProtector = opt.Val(game.Player2)
	plain.BattleProtector = opt.Val(game.Player2)
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.RemoveCounter{
		Amount:      game.Fixed(2),
		Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Battle}}),
		CounterKind: counter.Defense,
	}, &TurnLog{})

	events := countersRemovedEvents(g)
	if len(events) != 4 || events[0].SimultaneousID == 0 {
		t.Fatalf("counters removed events = %+v, want four simultaneous removals", events)
	}
	for _, event := range events[1:] {
		if event.SimultaneousID != events[0].SimultaneousID {
			t.Fatalf("counters removed events = %+v, want one shared SimultaneousID", events)
		}
	}
	defeatedOrWaitingSiege(t, g, engine, segovia, kylem)
	if siegeDefeatAbilities(g) != 2 {
		t.Fatalf("stack = %+v, want one defeat ability for each defeated Siege", g.Stack.Objects())
	}
	sources := map[game.ObjectID]game.PlayerID{}
	for _, obj := range g.Stack.Objects() {
		if obj.SiegeDefeat {
			sources[obj.SourceID] = obj.Controller
		}
	}
	if sources[segovia.ObjectID] != game.Player1 || sources[kylem.ObjectID] != game.Player2 {
		t.Fatalf("defeat abilities = %v, want Segovia's for Player1 and Kylem's for Player2", sources)
	}
	if _, ok := permanentByObjectID(g, survivor.ObjectID); !ok || survivor.Counters.Get(counter.Defense) != 1 {
		t.Fatal("Siege left with defense did not stay on the battlefield with 1 defense")
	}
	if !g.Players[game.Player1].Graveyard.Contains(plain.CardInstanceID) {
		t.Fatal("non-Siege battle with 0 defense was not put into its owner's graveyard")
	}
}

// TestMovingLastDefenseCountersTriggersSiege proves the source side of a
// counter move removes counters (CR 122.5): moving a Siege's last defense
// counters onto another permanent triggers it.
func TestMovingLastDefenseCountersTriggersSiege(t *testing.T) {
	for _, allKinds := range []bool{false, true} {
		g := mainPhaseGame(game.Player1)
		engine := choicesProgramEngine(t)
		siege := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
		siege.Counters.Add(counter.Defense, 2)
		siege.BattleProtector = opt.Val(game.Player2)
		destination := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Counter Vault", Types: []types.Card{types.Artifact},
		}})
		obj := &game.StackObject{
			ID:         g.IDGen.Next(),
			Kind:       game.StackSpell,
			Controller: game.Player2,
			Targets:    []game.Target{game.PermanentTarget(siege.ObjectID), game.PermanentTarget(destination.ObjectID)},
		}
		resolveInstruction(engine, g, obj, game.MoveCounters{
			Amount:      game.Fixed(2),
			Object:      game.TargetPermanentReference(1),
			CounterKind: counter.Defense,
			AllKinds:    allKinds,
			Source:      game.CounterSourceSpec{Kind: game.CounterSourceTarget, Object: game.TargetPermanentReference(0)},
		}, &TurnLog{})
		if siege.Counters.Get(counter.Defense) != 0 || destination.Counters.Get(counter.Defense) != 2 {
			t.Fatalf("allKinds=%v: Siege defense = %d, destination defense = %d; want the 2 counters moved",
				allKinds, siege.Counters.Get(counter.Defense), destination.Counters.Get(counter.Defense))
		}
		removed := lastCountersRemovedEvent(t, g)
		if removed.PermanentID != siege.ObjectID || removed.PreviousCounterAmount != 2 || removed.Amount != 2 || removed.Controller != game.Player2 {
			t.Fatalf("allKinds=%v: counters removed event = %+v, want 2 of 2 moved off the Siege by Player2", allKinds, removed)
		}
		defeatedOrWaitingSiege(t, g, engine, siege)
		siegeDefeatAbilityOnTop(t, g, siege)
	}
}

// TestRemoveDefenseCounterCostTriggersSiege proves an additional cost that
// removes a Siege's last defense counters triggers it (CR 601.2h, 310.12b).
func TestRemoveDefenseCounterCostTriggersSiege(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	siege := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	siege.Counters.Add(counter.Defense, 2)
	siege.BattleProtector = opt.Val(game.Player2)
	if _, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   siege,
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalRemoveCounter,
			Text:        "Remove two defense counters from this permanent",
			Amount:      2,
			CounterKind: counter.Defense,
		}},
	}); !ok {
		t.Fatal("could not pay the remove-defense-counters cost")
	}
	removed := lastCountersRemovedEvent(t, g)
	if removed.PermanentID != siege.ObjectID || removed.PreviousCounterAmount != 2 || removed.Amount != 2 || removed.Controller != game.Player1 {
		t.Fatalf("counters removed event = %+v, want 2 of 2 removed by Player1's cost", removed)
	}
	defeatedOrWaitingSiege(t, g, engine, siege)
	siegeDefeatAbilityOnTop(t, g, siege)
}

// TestRemovingOtherCounterKindFromSiegeDoesNotTrigger proves a removal of a
// non-defense kind publishes the event but never triggers a Siege, including a
// Siege that never had defense counters.
func TestRemovingOtherCounterKindFromSiegeDoesNotTrigger(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	siege := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	siege.Counters.Add(counter.Defense, 2)
	siege.Counters.Add(counter.Charge, 2)
	siege.BattleProtector = opt.Val(game.Player2)
	copied := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	copied.Counters.Add(counter.Charge, 1)
	copied.BattleProtector = opt.Val(game.Player2)
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.RemoveCounter{
		Amount:      game.Fixed(2),
		Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Battle}}),
		CounterKind: counter.Charge,
	}, &TurnLog{})
	if events := countersRemovedEvents(g); len(events) != 2 || events[0].CounterKind != counter.Charge {
		t.Fatalf("counters removed events = %+v, want the two charge removals", events)
	}
	engine.applyStateBasedActions(g)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if siegeDefeatAbilities(g) != 0 {
		t.Fatalf("stack = %+v, want no defeat ability", g.Stack.Objects())
	}
	if _, ok := permanentByObjectID(g, siege.ObjectID); !ok || siege.Counters.Get(counter.Defense) != 2 {
		t.Fatal("Siege with defense lost defense or left the battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(copied.CardInstanceID) {
		t.Fatal("0-defense Siege without a defeat trigger was not put into its owner's graveyard")
	}
}
