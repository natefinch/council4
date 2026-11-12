package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Effects target suspended cards to remove or put time counters (#3790).
//
// Source review: official WotC Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829:
// 702.62a (suspend's cast ability triggers "When the last time counter is
// removed from this card, if it's exiled"), 702.62b (a card is "suspended" if
// it's in exile, has suspend and has a time counter), 115.2 (only permanents
// are legal targets unless the ability says it can target an object in
// another zone), 603.3c/700.2b (a mode with no legal targets can't be chosen),
// 608.2b (a target no longer in the zone it was targeted in is illegal and
// isn't affected) and 400.7 (a card that changes zones is a new object).
// Shivan Sand-Mage, oracle_id 3afe854a-7daa-4745-a5d9-04879e5d940a, pinned
// corpus and https://api.scryfall.com/cards/oracle/3afe854a-7daa-4745-a5d9-04879e5d940a
// rulings (checked 2026-11-12): "When this creature enters, choose one — •
// Remove two time counters from target permanent or suspended card. • Put two
// time counters on target permanent with a time counter on it or suspended
// card. Suspend 4—{R}". Rulings: 2007-05-01 "If there are no suspended cards
// and no permanents with time counters on them, you can't choose the second
// mode"; 2024-02-02 the last time counter being removed triggers the cast
// ability no matter what removed it, and if that ability is countered the
// card stays exiled with no time counters. Ancestral Vision, oracle_id
// 9728dec9-d482-4c7a-8cdc-44d010dc878d. Stifle, oracle_id
// b3b00911-ece7-4484-bc36-f211ce72b6cc: "Counter target activated or
// triggered ability."
//
// Shivan Sand-Mage itself is not registered by this Engine change; the
// fixture below declares its entry ability with the shared target and counter
// operations so a later card unit can author the card.

const (
	sandMageRemoveMode = 0
	sandMagePutMode    = 1
)

func shivanSandMageFixture() *game.CardDef {
	removeTarget := game.TargetSpec{
		MinTargets: 1, MaxTargets: 1,
		Allow:      game.TargetAllowPermanent | game.TargetAllowCard,
		TargetZone: zone.Exile, SuspendedCard: true,
	}
	putTarget := removeTarget
	putTarget.Selection = opt.Val(game.Selection{MatchCounter: true, RequiredCounter: counter.Time})
	pt := game.PT{Value: 3}
	toughness := game.PT{Value: 2}
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Shivan Sand-Mage", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R, cost.R}),
		Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Lizard, types.Shaman},
		Colors: []color.Color{color.Red}, Power: opt.Val(pt), Toughness: opt.Val(toughness),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.SuspendKeyword{Cost: cost.Mana{cost.R}, TimeCounters: 4}},
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
			}},
			Content: game.AbilityContent{MinModes: 1, MaxModes: 1, Modes: []game.Mode{
				{
					Text:    "Remove two time counters from target permanent or suspended card.",
					Targets: []game.TargetSpec{removeTarget},
					Sequence: []game.Instruction{{Primitive: game.RemoveCounter{
						Amount: game.Fixed(2), Object: game.TargetObjectReference(0), CounterKind: counter.Time,
					}}},
				},
				{
					Text:    "Put two time counters on target permanent with a time counter on it or suspended card.",
					Targets: []game.TargetSpec{putTarget},
					Sequence: []game.Instruction{{Primitive: game.AddCounter{
						Amount: game.Fixed(2), Object: game.TargetObjectReference(0), CounterKind: counter.Time,
					}}},
				},
			}},
		}},
	}}
}

// sandMagePolicy chooses the Sand-Mage's mode and target, recording the modes
// and targets it was offered, and otherwise takes default choices.
type sandMagePolicy struct {
	simPassPolicy

	mode         int
	target       game.Target
	offeredModes []int
	offered      []game.Target
}

func (p *sandMagePolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceModal:
		p.offeredModes = nil
		for _, option := range request.Options {
			p.offeredModes = append(p.offeredModes, option.Index)
		}
		for _, option := range request.Options {
			if option.Index == p.mode {
				return []int{option.Index}
			}
		}
	case game.ChoiceTarget:
		p.offered = nil
		chosen := -1
		for _, option := range request.Options {
			if len(option.Targets) != 1 {
				continue
			}
			p.offered = append(p.offered, option.Targets[0])
			if sameTargetObject(option.Targets[0], p.target) {
				chosen = option.Index
			}
		}
		if chosen >= 0 {
			return []int{chosen}
		}
	}
	return request.DefaultSelection
}

func offeredTarget(offered []game.Target, target game.Target) bool {
	return slices.ContainsFunc(offered, func(candidate game.Target) bool { return sameTargetObject(candidate, target) })
}

// castSandMageEntry casts the Shivan Sand-Mage fixture in Player 1's main
// phase, resolves it and puts its entry ability on the stack with the
// policy's mode and target.
func castSandMageEntry(t *testing.T, engine *Engine, g *game.Game, policy *sandMagePolicy) (*game.Permanent, *game.StackObject) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, shivanSandMageFixture())
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)
	agents := suspendAgents(policy)
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Shivan Sand-Mage failed")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	sandMage := permanentByCardID(g, cardID)
	if sandMage == nil {
		t.Fatal("Shivan Sand-Mage did not enter the battlefield")
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility || obj.SourceCardID != cardID {
		return sandMage, nil
	}
	return sandMage, obj
}

func requireSandMageEntry(t *testing.T, obj *game.StackObject, mode int, target game.Target) {
	t.Helper()
	if obj == nil {
		t.Fatal("Shivan Sand-Mage's entry ability is not on the stack")
	}
	if !slices.Equal(obj.ChosenModes, []int{mode}) || len(obj.Targets) != 1 || !sameTargetObject(obj.Targets[0], target) {
		t.Fatalf("entry ability modes=%v targets=%v, want mode %d targeting %v", obj.ChosenModes, obj.Targets, mode, target)
	}
}

// TestShivanSandMageRemovesSuspendedCardTimeCounters plays the Sand-Mage's
// "remove two time counters from target ... suspended card" mode in Player 1's
// main phase. Removing the last counter triggers suspend's cast ability on the
// stack, which casts Ancestral Vision for free outside the upkeep, or leaves it
// exiled with no counters when Stifle counters that ability.
func TestShivanSandMageRemovesSuspendedCardTimeCounters(t *testing.T) {
	for _, tc := range []struct {
		name     string
		counters int
		stifle   bool
	}{
		{name: "two counters", counters: 2},
		{name: "one counter", counters: 1},
		{name: "stifled cast trigger", counters: 2, stifle: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			vision := addSuspendedAncestralVision(t, g, game.Player1, tc.counters)
			visionTarget := game.CardTarget(vision)
			policy := &sandMagePolicy{mode: sandMageRemoveMode, target: visionTarget}
			_, entry := castSandMageEntry(t, engine, g, policy)
			requireSandMageEntry(t, entry, sandMageRemoveMode, visionTarget)
			if got := g.ExileCounterCount(vision, counter.Time); got != tc.counters {
				t.Fatalf("time counters = %d before the entry ability resolved, want %d", got, tc.counters)
			}

			agents := suspendAgents(policy)
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			requireExileTimeCounterEvent(t, g, game.EventCountersRemoved, vision, tc.counters, tc.counters)
			if g.ExileCounterCount(vision, counter.Time) != 0 || cardIsSuspended(g, vision) {
				t.Fatal("the entry ability did not remove the last time counter")
			}
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			castTrigger := requireSuspendCastTrigger(t, g, vision)
			if g.Turn.Phase != game.PhasePrecombatMain {
				t.Fatalf("phase = %v, want the cast trigger in the main phase", g.Turn.Phase)
			}

			if tc.stifle {
				stifleID := addCardToHand(g, game.Player2, registryCard(t, "Stifle"))
				g.Players[game.Player2].ManaPool.Add(mana.U, 1)
				g.Turn.PriorityPlayer = game.Player2
				if !engine.applyAction(g, game.Player2, action.CastSpell(stifleID, []game.Target{game.StackObjectTarget(castTrigger.ID)}, 0, nil)) {
					t.Fatal("Stifle could not target the suspend cast trigger")
				}
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				requireSuspendNotCast(t, g, vision)
				return
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			spell, _ := requireSuspendCast(t, g, vision)
			if spell.Controller != game.Player1 {
				t.Fatalf("Ancestral Vision controller = %v, want its owner", spell.Controller)
			}
		})
	}
}

// TestShivanSandMagePutsTimeCountersOnSuspendedCard plays the "put two time
// counters on target ... suspended card" mode: the card gets two more time
// counters, and Player 1's next upkeep removes only one of them.
func TestShivanSandMagePutsTimeCountersOnSuspendedCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	vision := addSuspendedAncestralVision(t, g, game.Player1, 1)
	visionTarget := game.CardTarget(vision)
	policy := &sandMagePolicy{mode: sandMagePutMode, target: visionTarget}
	_, entry := castSandMageEntry(t, engine, g, policy)
	requireSandMageEntry(t, entry, sandMagePutMode, visionTarget)

	agents := suspendAgents(policy)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	requireExileTimeCounterEvent(t, g, game.EventCountersAdded, vision, 1, 2)
	if got := g.ExileCounterCount(vision, counter.Time); got != 3 {
		t.Fatalf("time counters = %d, want 3 after putting two on", got)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no trigger from putting time counters on", g.Stack.Size())
	}

	engine.runSuspendUpkeep(g, game.Player1, agents)
	if got := g.ExileCounterCount(vision, counter.Time); got != 2 || !cardIsSuspended(g, vision) {
		t.Fatalf("time counters = %d after the upkeep, want 2 with the card still suspended", got)
	}
	if len(suspendCastEvents(g, vision)) != 0 {
		t.Fatal("Ancestral Vision was cast with time counters remaining")
	}
}

// TestShivanSandMageTargetLegality checks which objects the entry ability may
// target. Only exiled cards with suspend and a time counter are legal card
// targets: an exiled card without suspend, a suspend card in exile without
// time counters and a suspend card in a graveyard aren't. The second mode
// also needs a permanent with a time counter or a suspended card, so it can't
// be chosen without one, and a permanent with time counters is still a legal
// target whose counters the first mode removes.
func TestShivanSandMageTargetLegality(t *testing.T) {
	t.Run("suspended cards only", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := choicesProgramEngine(t)
		vision := addSuspendedAncestralVision(t, g, game.Player2, 3)
		noSuspend := addSuspendedCard(g, game.Player1, vanillaCreature("Exiled Bear", 2, 2), 2)
		noCounters := addSuspendedAncestralVision(t, g, game.Player1, 0)
		graveyard := addCardToGraveyard(g, game.Player1, registryCard(t, "Ancestral Vision"))
		clock := addCombatPermanent(g, game.Player2, vanillaCreature("Clock", 1, 1))
		clock.Counters.Add(counter.Time, 1)
		policy := &sandMagePolicy{mode: sandMagePutMode, target: game.CardTarget(vision)}
		_, entry := castSandMageEntry(t, engine, g, policy)
		requireSandMageEntry(t, entry, sandMagePutMode, game.CardTarget(vision))
		if !slices.Equal(policy.offeredModes, []int{sandMageRemoveMode, sandMagePutMode}) {
			t.Fatalf("offered modes = %v, want both", policy.offeredModes)
		}
		if len(policy.offered) != 2 ||
			!offeredTarget(policy.offered, game.CardTarget(vision)) ||
			!offeredTarget(policy.offered, game.PermanentTarget(clock.ObjectID)) {
			t.Fatalf("put-mode targets = %v, want the suspended Ancestral Vision and the permanent with a time counter", policy.offered)
		}
		for _, illegal := range []id.ID{noSuspend, noCounters, graveyard} {
			if offeredTarget(policy.offered, game.CardTarget(illegal)) {
				t.Fatalf("card %v without suspend, time counters or exile was offered", illegal)
			}
		}

		engine.resolveTopOfStackWithChoices(g, suspendAgents(policy), &TurnLog{})
		if got := g.ExileCounterCount(vision, counter.Time); got != 5 {
			t.Fatalf("opponent's suspended card time counters = %d, want 5", got)
		}
	})

	t.Run("second mode needs a time counter", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := choicesProgramEngine(t)
		addSuspendedAncestralVision(t, g, game.Player1, 0)
		bear := addCombatPermanent(g, game.Player2, vanillaCreature("Bear", 2, 2))
		policy := &sandMagePolicy{mode: sandMagePutMode, target: game.PermanentTarget(bear.ObjectID)}
		_, entry := castSandMageEntry(t, engine, g, policy)
		if !slices.Equal(policy.offeredModes, []int{sandMageRemoveMode}) {
			t.Fatalf("offered modes = %v, want only the remove mode", policy.offeredModes)
		}
		requireSandMageEntry(t, entry, sandMageRemoveMode, entry.Targets[0])
		if entry.Targets[0].Kind != game.TargetPermanent {
			t.Fatalf("remove-mode target = %v, want a permanent", entry.Targets[0])
		}
	})

	t.Run("permanent with time counters", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := choicesProgramEngine(t)
		vision := addSuspendedAncestralVision(t, g, game.Player1, 2)
		clock := addCombatPermanent(g, game.Player2, vanillaCreature("Clock", 1, 1))
		clock.Counters.Add(counter.Time, 3)
		policy := &sandMagePolicy{mode: sandMageRemoveMode, target: game.PermanentTarget(clock.ObjectID)}
		sandMage, entry := castSandMageEntry(t, engine, g, policy)
		requireSandMageEntry(t, entry, sandMageRemoveMode, game.PermanentTarget(clock.ObjectID))
		for _, want := range []game.Target{game.PermanentTarget(sandMage.ObjectID), game.CardTarget(vision)} {
			if !offeredTarget(policy.offered, want) {
				t.Fatalf("remove-mode targets = %v, want any permanent and the suspended card", policy.offered)
			}
		}
		engine.resolveTopOfStackWithChoices(g, suspendAgents(policy), &TurnLog{})
		if got := clock.Counters.Get(counter.Time); got != 1 {
			t.Fatalf("permanent time counters = %d, want 1", got)
		}
		if got := g.ExileCounterCount(vision, counter.Time); got != 2 {
			t.Fatalf("untargeted suspended card time counters = %d, want 2", got)
		}
	})
}

// TestShivanSandMageTargetLeavesExile moves the targeted suspended card out of
// exile and back before the entry ability resolves. The returned card is a new
// object (CR 400.7), so the target is illegal (CR 608.2b): even suspended
// again with two time counters, the ability does nothing to it.
func TestShivanSandMageTargetLeavesExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	vision := addSuspendedAncestralVision(t, g, game.Player1, 2)
	policy := &sandMagePolicy{mode: sandMageRemoveMode, target: game.CardTarget(vision)}
	_, entry := castSandMageEntry(t, engine, g, policy)
	requireSandMageEntry(t, entry, sandMageRemoveMode, game.CardTarget(vision))
	if !moveCardBetweenZones(g, game.Player1, vision, zone.Exile, zone.Graveyard) ||
		!moveCardBetweenZones(g, game.Player1, vision, zone.Graveyard, zone.Exile) {
		t.Fatal("moving the targeted card out of exile and back failed")
	}
	g.AddExileCounter(vision, counter.Time, 2)

	agents := suspendAgents(policy)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if got := g.ExileCounterCount(vision, counter.Time); got != 2 || suspendUpkeepRemovals(g, vision) != 0 {
		t.Fatalf("returned card time counters = %d, want the new object untouched", got)
	}
	if g.Stack.Size() != 0 || len(suspendCastEvents(g, vision)) != 0 {
		t.Fatal("the illegal target's suspend cast ability triggered")
	}
}

func mustCard(t *testing.T, g *game.Game, cardID id.ID) *game.CardInstance {
	t.Helper()
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatalf("card %v is missing", cardID)
	}
	return card
}
