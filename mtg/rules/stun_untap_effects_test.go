package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestStunUntapResolvingTapCounterUntapSequence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Target Vehicle", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Vehicle},
	}})
	ref := game.TargetPermanentReference(0)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Tap{Object: ref}},
		{Primitive: game.AddCounter{Object: ref, CounterKind: counter.Stun, Amount: game.Fixed(3)}},
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !target.Tapped || target.Counters.Get(counter.Stun) != 3 {
		t.Fatalf("tap/counter resolution = tapped %v, stun %d; want true, 3", target.Tapped, target.Counters.Get(counter.Stun))
	}
	start := len(g.Events)
	addEffectSpellToStack(g, game.Player1, game.Untap{Object: ref}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !target.Tapped || target.Counters.Get(counter.Stun) != 2 {
		t.Fatalf("resolving untap = tapped %v, stun %d; want true, 2", target.Tapped, target.Counters.Get(counter.Stun))
	}
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventPermanentUntapped && event.PermanentID == target.ObjectID {
			t.Fatal("replaced untap emitted a permanent-untapped event")
		}
	}
}

func TestStunUntapEffectRoutes(t *testing.T) {
	group := game.PlayerControlledGroup(game.ControllerReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}})
	routes := []struct {
		name      string
		primitive game.Primitive
		answer    []int
	}{
		{"object", game.Untap{Object: game.TargetPermanentReference(0)}, nil},
		{"group", game.Untap{Group: group}, nil},
		{"choose-one", game.Untap{Group: group, ChooseOne: true}, []int{0}},
		{"choose-up-to", game.Untap{Group: group, ChooseUpTo: true, Amount: game.Fixed(1)}, []int{0}},
		{"tap-or-untap", game.TapOrUntap{Object: game.TargetPermanentReference(0)}, []int{1}},
	}
	cases := []struct {
		name   string
		tapped bool
		stun   int
	}{
		{"zero", true, 0},
		{"one", true, 1},
		{"two", true, 2},
		{"three", true, 3},
		{"already-untapped", false, 2},
	}
	for _, route := range routes {
		for _, tc := range cases {
			t.Run(route.name+"/"+tc.name, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := addCreaturePermanent(g, game.Player1)
				target.Tapped = tc.tapped
				target.Counters.Add(counter.Stun, tc.stun)
				target.Counters.Add(counter.Charge, 2)
				cardID, objectID, owner, controller := target.CardInstanceID, target.ObjectID, target.Owner, target.Controller
				obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
				agents := [game.NumPlayers]PlayerAgent{game.Player1: scriptedChoiceAgent{answer: route.answer}}
				receipt, available := NewEngine(nil).resolveInstructionSequenceReceipt(g, obj, []game.Instruction{
					{Primitive: route.primitive, PublishResult: "untap"},
				}, agents, &TurnLog{}, "untap")
				untapped := tc.tapped && tc.stun == 0
				wantStun := tc.stun
				if tc.tapped && tc.stun > 0 {
					wantStun--
				}
				if target.Tapped != (tc.tapped && !untapped) || target.Counters.Get(counter.Stun) != wantStun {
					t.Fatalf("state = tapped %v, stun %d; want tapped %v, stun %d", target.Tapped, target.Counters.Get(counter.Stun), tc.tapped && !untapped, wantStun)
				}
				if target.Counters.Get(counter.Charge) != 2 {
					t.Fatal("untap changed unrelated counters")
				}
				if target.CardInstanceID != cardID || target.ObjectID != objectID || target.Owner != owner || target.Controller != controller {
					t.Fatal("untap changed physical permanent identity")
				}
				if !available || !receipt.Accepted || receipt.Succeeded != untapped {
					t.Fatalf("receipt = %+v, available %v; want accepted, succeeded %v", receipt, available, untapped)
				}
				events := stunUntappedEvents(g)
				if untapped {
					if len(events) != 1 || events[0].PermanentID != target.ObjectID {
						t.Fatalf("untapped events = %+v, want target exactly once", events)
					}
				} else if len(events) != 0 {
					t.Fatalf("replaced/no-op untap emitted events: %+v", events)
				}
			})
		}
	}
}

func stunUntappedEvents(g *game.Game) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentUntapped {
			events = append(events, event)
		}
	}
	return events
}

func TestStunUntapMixedSimultaneousGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stunned := addCreaturePermanent(g, game.Player1)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player2)
	untapped := addCreaturePermanent(g, game.Player2)
	stunned.Tapped, first.Tapped, second.Tapped = true, true, true
	stunned.Counters.Add(counter.Stun, 2)
	untapped.Counters.Add(counter.Stun, 2)
	obj := &game.StackObject{Controller: game.Player1}
	receipt, available := NewEngine(nil).resolveInstructionSequenceReceipt(g, obj, []game.Instruction{{
		Primitive:     game.Untap{Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})},
		PublishResult: "untap",
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}, "untap")
	if !stunned.Tapped || stunned.Counters.Get(counter.Stun) != 1 || untapped.Counters.Get(counter.Stun) != 2 || first.Tapped || second.Tapped {
		t.Fatal("mixed group did not replace only the stunned untap")
	}
	if !available || !receipt.Succeeded {
		t.Fatalf("mixed group receipt = %+v, available %v; want successful actual untaps", receipt, available)
	}
	events := stunUntappedEvents(g)
	if len(events) != 2 || events[0].PermanentID != first.ObjectID || events[1].PermanentID != second.ObjectID ||
		events[0].Controller != game.Player1 || events[1].Controller != game.Player2 ||
		events[0].SimultaneousID == 0 || events[0].SimultaneousID != events[1].SimultaneousID {
		t.Fatalf("mixed group's untapped events = %+v, want two physical subjects in one simultaneous batch", events)
	}
}

func TestStunUntapChoiceMaySelectNone(t *testing.T) {
	for _, amount := range []int{0, 1} {
		t.Run(strconv.Itoa(amount), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCreaturePermanent(g, game.Player1)
			target.Tapped = true
			target.Counters.Add(counter.Stun, 2)
			obj := &game.StackObject{Controller: game.Player1}
			receipt, available := NewEngine(nil).resolveInstructionSequenceReceipt(g, obj, []game.Instruction{{
				Primitive: game.Untap{
					Group:      game.PlayerControlledGroup(game.ControllerReference(), game.Selection{}),
					ChooseUpTo: true, Amount: game.Fixed(amount),
				},
				PublishResult: "untap",
			}}, [game.NumPlayers]PlayerAgent{game.Player1: scriptedChoiceAgent{answer: []int{}}}, &TurnLog{}, "untap")
			if !target.Tapped || target.Counters.Get(counter.Stun) != 2 || len(stunUntappedEvents(g)) != 0 ||
				!available || !receipt.Accepted || receipt.Succeeded {
				t.Fatalf("no chosen untap mutated state or reported success: %+v, available %v", receipt, available)
			}
		})
	}
}

func TestStunUntapDoesNotTriggerUntilActuallyUntapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventPermanentUntapped, Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}, nil)
	source.Tapped = true
	source.Counters.Add(counter.Stun, 1)
	obj := &game.StackObject{Controller: game.Player1, Kind: game.StackActivatedAbility, SourceID: source.ObjectID}
	instr := &game.Instruction{Primitive: game.Untap{Object: game.SourcePermanentReference()}}
	engine.resolveInstructionWithChoices(g, obj, instr, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("stun-replaced untap put an untapped trigger on the stack")
	}
	engine.resolveInstructionWithChoices(g, obj, instr, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("actual untap did not put its trigger on the stack")
	}
	life := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+1 {
		t.Fatal("actual untap trigger did not resolve")
	}
}

func TestStunUntapChosenGroupIsSimultaneous(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stunned := addCreaturePermanent(g, game.Player1)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player2)
	notChosen := addCreaturePermanent(g, game.Player2)
	for _, permanent := range []*game.Permanent{stunned, first, second, notChosen} {
		permanent.Tapped = true
	}
	stunned.Counters.Add(counter.Stun, 2)
	notChosen.Counters.Add(counter.Stun, 2)
	obj := &game.StackObject{Controller: game.Player3}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.Untap{
		Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		ChooseUpTo: true, Amount: game.Fixed(3),
	}}, [game.NumPlayers]PlayerAgent{game.Player3: scriptedChoiceAgent{answer: []int{0, 1, 2}}}, &TurnLog{})
	if !stunned.Tapped || stunned.Counters.Get(counter.Stun) != 1 ||
		first.Tapped || second.Tapped || !notChosen.Tapped || notChosen.Counters.Get(counter.Stun) != 2 {
		t.Fatal("choice group changed an unchosen object or mishandled replacement")
	}
	events := stunUntappedEvents(g)
	if len(events) != 2 || events[0].PermanentID != first.ObjectID || events[1].PermanentID != second.ObjectID ||
		events[0].SimultaneousID == 0 || events[0].SimultaneousID != events[1].SimultaneousID {
		t.Fatalf("chosen-group events = %+v, want two actual untaps in one batch", events)
	}
}

func TestStunUntapDuplicateSubjectsAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCreaturePermanent(g, game.Player1)
	target.Tapped = true
	target.Counters.Add(counter.Stun, 2)
	clone := g.Clone()
	cloned := clone.Battlefield[0]
	if setPermanentsTappedSimultaneously(clone, []*game.Permanent{nil, cloned, cloned}, false) {
		t.Fatal("all-replaced group reported an actual untap")
	}
	if !cloned.Tapped || cloned.Counters.Get(counter.Stun) != 1 || len(stunUntappedEvents(clone)) != 0 {
		t.Fatal("one batch applied stun replacement more than once to the same physical object")
	}
	if !target.Tapped || target.Counters.Get(counter.Stun) != 2 || len(stunUntappedEvents(g)) != 0 {
		t.Fatal("replacing an untap in a clone changed the original")
	}
}
