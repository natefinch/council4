package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestRemoveAllOfKindFromAllPermanentsActivation plays the Oblivion Stone
// shape "{T}: Remove all fate counters from all permanents. You gain 1 life for
// each counter removed this way." through activation and resolution. Every fate
// counter on every permanent, whatever its controller or count, is removed as
// one simultaneous action (CR 122.8 removal; CR 608.2c "this way" reads the
// actual result), other kinds stay, and the follow-on reads the total removed.
func TestRemoveAllOfKindFromAllPermanentsActivation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	const removed game.ResultKey = "fate-removed"
	stone := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Fate Stone",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
			Content: game.Mode{Sequence: []game.Instruction{
				{
					Primitive: game.RemoveCounter{
						Group:       game.BattlefieldGroup(game.Selection{}),
						CounterKind: counter.Fate,
						AllOfKind:   true,
					},
					PublishResult: removed,
				},
				{Primitive: game.GainLife{
					Player: game.ControllerReference(),
					Amount: game.Dynamic(game.DynamicAmount{
						Kind: game.DynamicAmountPreviousEffectResult, ResultKey: removed, Multiplier: 1,
					}),
				}},
			}}.Ability(),
		}},
	}})
	mine := addCombatPermanent(g, game.Player1, creatureDef("Fated Ally"))
	mine.Counters.Add(counter.Fate, 2)
	mine.Counters.Add(counter.Charge, 1)
	theirs := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Fated Relic", Types: []types.Card{types.Artifact},
	}})
	theirs.Counters.Add(counter.Fate, 1)
	bystander := addCombatPermanent(g, game.Player3, creatureDef("Unfated"))
	bystander.Counters.Add(counter.PlusOnePlusOne, 2)
	eventsBefore := len(g.Events)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(stone.ObjectID, 0, nil, 0)) {
		t.Fatal("Fate Stone activation failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := mine.Counters.Get(counter.Fate) + theirs.Counters.Get(counter.Fate); got != 0 {
		t.Fatalf("fate counters left = %d, want every fate counter removed", got)
	}
	if mine.Counters.Get(counter.Charge) != 1 || bystander.Counters.Get(counter.PlusOnePlusOne) != 2 {
		t.Fatalf("other kinds changed: charge=%d +1/+1=%d", mine.Counters.Get(counter.Charge), bystander.Counters.Get(counter.PlusOnePlusOne))
	}
	if life := g.Players[game.Player1].Life; life != 43 {
		t.Fatalf("life = %d, want 43 (1 for each of 3 fate counters removed)", life)
	}
	var events []game.Event
	for _, event := range g.Events[eventsBefore:] {
		if event.Kind == game.EventCountersRemoved {
			events = append(events, event)
		}
	}
	if len(events) != 2 {
		t.Fatalf("counters removed events = %+v, want one per permanent that had fate counters", events)
	}
	previous := map[game.ObjectID]int{}
	for _, event := range events {
		if event.CounterKind != counter.Fate || event.Amount != event.PreviousCounterAmount ||
			event.SimultaneousID == 0 || event.SimultaneousID != events[0].SimultaneousID || event.Controller != game.Player1 {
			t.Fatalf("counters removed event = %+v, want all fate counters removed simultaneously by Player1", event)
		}
		previous[event.PermanentID] = event.PreviousCounterAmount
	}
	if previous[mine.ObjectID] != 2 || previous[theirs.ObjectID] != 1 {
		t.Fatalf("removed amounts = %v, want 2 from the ally and 1 from the relic", previous)
	}
}

// TestRemoveAllOfKindFromSourceGatesOnRemoval plays the Bounty of the Luxa
// shape: "Remove all flood counters from this enchantment. If no counters were
// removed this way, put a flood counter on it. Otherwise, you gain life equal
// to the number removed." Removing none fails the removal; removing several
// removes every one regardless of count and reports that count.
func TestRemoveAllOfKindFromSourceGatesOnRemoval(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Flood Bounty", Types: []types.Card{types.Enchantment},
	}})
	const removed game.ResultKey = "flood-removed"
	sequence := []game.Instruction{
		{
			Primitive: game.RemoveCounter{
				Object:      game.SourcePermanentReference(),
				CounterKind: counter.Flood,
				AllOfKind:   true,
			},
			PublishResult: removed,
		},
		{
			Primitive:  game.AddCounter{Amount: game.Fixed(1), Object: game.SourcePermanentReference(), CounterKind: counter.Flood},
			ResultGate: opt.Val(game.InstructionResultGate{Key: removed, Succeeded: game.TriFalse}),
		},
		{
			Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
				Kind: game.DynamicAmountPreviousEffectResult, ResultKey: removed, Multiplier: 1,
			})},
			ResultGate: opt.Val(game.InstructionResultGate{Key: removed, Succeeded: game.TriTrue}),
		},
	}
	if err := game.ValidateInstructionSequence(sequence, nil); err != nil {
		t.Fatalf("sequence rejected: %v", err)
	}
	resolve := func() {
		obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: source.ObjectID, Controller: game.Player1}
		engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}

	resolve()
	if source.Counters.Get(counter.Flood) != 1 || g.Players[game.Player1].Life != 40 {
		t.Fatalf("no-counter resolution: flood=%d life=%d, want a flood counter put on and no life", source.Counters.Get(counter.Flood), g.Players[game.Player1].Life)
	}
	source.Counters.Add(counter.Flood, 2)
	source.Counters.Add(counter.Charge, 1)
	resolve()
	if source.Counters.Get(counter.Flood) != 0 || source.Counters.Get(counter.Charge) != 1 || g.Players[game.Player1].Life != 43 {
		t.Fatalf("three-counter resolution: flood=%d charge=%d life=%d, want all 3 flood removed, charge kept, 3 life",
			source.Counters.Get(counter.Flood), source.Counters.Get(counter.Charge), g.Players[game.Player1].Life)
	}
}

// TestRemoveAllOfKindFromTargetedExiledCard removes every counter of the named
// kind from a targeted card in exile, leaving its other kinds.
func TestRemoveAllOfKindFromTargetedExiledCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToExile(g, game.Player2, creatureDef("Suspended Beast"))
	g.AddExileCounter(cardID, counter.Time, 4)
	g.AddExileCounter(cardID, counter.Charge, 1)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
		Targets: []game.Target{game.CardTarget(cardID)},
	}
	resolveInstruction(engine, g, obj, game.RemoveCounter{
		Object:      game.TargetObjectReference(0),
		CounterKind: counter.Time,
		AllOfKind:   true,
	}, &TurnLog{})
	if g.ExileCounterCount(cardID, counter.Time) != 0 || g.ExileCounterCount(cardID, counter.Charge) != 1 {
		t.Fatalf("exiled counters: time=%d charge=%d, want all time removed and charge kept",
			g.ExileCounterCount(cardID, counter.Time), g.ExileCounterCount(cardID, counter.Charge))
	}
	if removed := lastCountersRemovedEvent(t, g); removed.CardID != cardID || removed.Amount != 4 || removed.PreviousCounterAmount != 4 {
		t.Fatalf("counters removed event = %+v, want all 4 time counters", removed)
	}
}

func TestRemoveAllOfKindValidation(t *testing.T) {
	targets := []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "permanent"}}
	for name, prim := range map[string]game.RemoveCounter{
		"amount":     {Object: game.TargetPermanentReference(0), CounterKind: counter.Fate, AllOfKind: true, Amount: game.Fixed(1)},
		"all kinds":  {Object: game.TargetPermanentReference(0), CounterKind: counter.Fate, AllOfKind: true, AllKinds: true},
		"choose":     {Object: game.TargetPermanentReference(0), CounterKind: counter.Fate, AllOfKind: true, ChooseKind: true},
		"no subject": {CounterKind: counter.Fate, AllOfKind: true},
	} {
		if err := game.ValidateInstructionSequence([]game.Instruction{{Primitive: prim}}, targets); err == nil {
			t.Errorf("%s: invalid remove-all-of-kind accepted", name)
		}
	}
	valid := game.RemoveCounter{Object: game.TargetPermanentReference(0), CounterKind: counter.Fate, AllOfKind: true}
	if err := game.ValidateInstructionSequence([]game.Instruction{{Primitive: valid}}, targets); err != nil {
		t.Fatalf("valid remove-all-of-kind rejected: %v", err)
	}
}
