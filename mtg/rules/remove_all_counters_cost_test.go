package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3882 ("Remove all <kind> counters from this"
// activation cost with the removed count available to the effect), verified
// against the pinned corpus, Scryfall rulings and the Comprehensive Rules
// effective 2025-09-19.
//
// CR 118.3: a cost can't be paid without the resources to pay it fully.
// CR 601.2h / 602.2b: the activation cost is paid in full when the ability is
// activated; "remove all" removes however many counters there are, including
// none. CR 707.10: a copy of an ability copies all decisions made for it,
// including additional costs, and an effect of the copy referring to what paid
// the costs uses what paid the original's.
//
// Molten Hydra (oracle cd5bc95d-4915-4c5a-9b57-2730adf69797): "{1}{R}{R}: Put a
// +1/+1 counter on this creature. {T}, Remove all +1/+1 counters from this
// creature: It deals damage to any target equal to the number of +1/+1
// counters removed this way." Rulings: none.
// Sage of Hours (oracle bdbf70da-af38-4827-9b8b-39f553ebd34a): "Remove all
// +1/+1 counters from this creature: For each five counters removed this way,
// take an extra turn after this one." Ruling 2014-04-26: you must remove all
// +1/+1 counters even if the number isn't a multiple of five; removing twelve
// gives two extra turns and removing three gives none.
//
// The probe below uses a Molten Hydra-shaped inline ability with an added {R}
// so combined mana, {T} and remove-all payment are exercised together; cards are
// admitted by a later card unit.

const removeAllCounterKey = "removed counters"

func removeAllCounterProbe(divisor int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Remove-All Probe",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 5}),
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:     "{R}, {T}, Remove all +1/+1 counters from this creature: It deals damage to any target equal to the number of +1/+1 counters removed this way.",
			ManaCost: opt.Val(cost.Mana{cost.R}),
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalTap},
				{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.PlusOnePlusOne, RemoveAll: true, SubjectKey: removeAllCounterKey},
			},
			ZoneOfFunction: zone.Battlefield,
			Content: game.AbilityContent{Modes: []game.Mode{{
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Constraint: "any target",
					Allow:     game.TargetAllowPermanent | game.TargetAllowPlayer,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind: game.DynamicAmountPaidCostCounters, CostKey: removeAllCounterKey, Divisor: divisor,
					}),
					Recipient:    game.AnyTargetDamageRecipient(0),
					DamageSource: opt.Val(game.SourcePermanentReference()),
				}}},
			}}, MinModes: 1, MaxModes: 1},
		}},
	}}
}

func activateRemoveAllProbe(t *testing.T, g *game.Game, e *Engine, source *game.Permanent) *game.StackObject {
	t.Helper()
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.PlayerTarget(game.Player2)}, 0)) {
		t.Fatal("remove-all activation rejected")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackActivatedAbility || obj.SourceID != source.ObjectID {
		t.Fatalf("stack top = %+v, want the remove-all ability", obj)
	}
	return obj
}

func wantPaidCounters(t *testing.T, obj *game.StackObject, want int) {
	t.Helper()
	if len(obj.PaidCostSubjects) != 1 {
		t.Fatalf("paid cost subjects = %+v, want one counter removal", obj.PaidCostSubjects)
	}
	got := obj.PaidCostSubjects[0]
	if got.Key != removeAllCounterKey || got.Kind != game.PaidCostCounterRemoval || got.Amount != want {
		t.Fatalf("paid cost subject = %+v, want %d counters removed under %q", got, want, removeAllCounterKey)
	}
}

// TestRemoveAllCountersCostLocksPaidCountForCopiesAndClones activates the
// remove-all ability with three +1/+1 and two charge counters, then puts two new
// +1/+1 counters on the source and copies the ability. The original, its copy
// and a cloned game each deal exactly the three damage paid; the charge counters
// and the later +1/+1 counters stay (CR 601.2h, 602.2b, 707.10).
func TestRemoveAllCountersCostLocksPaidCountForCopiesAndClones(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, removeAllCounterProbe(0))
	source.Counters.Add(counter.PlusOnePlusOne, 3)
	source.Counters.Add(counter.Charge, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	life := g.Players[game.Player2].Life

	original := activateRemoveAllProbe(t, g, e, source)
	if !source.Tapped || source.Counters.Get(counter.PlusOnePlusOne) != 0 || source.Counters.Get(counter.Charge) != 2 {
		t.Fatalf("source tapped=%v counters=%v, want tapped with every +1/+1 counter and no charge counter removed", source.Tapped, source.Counters)
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("the {R} part of the activation cost was not paid")
	}
	if removed := lastCountersRemovedEvent(t, g); removed.PermanentID != source.ObjectID ||
		removed.CounterKind != counter.PlusOnePlusOne || removed.Amount != 3 {
		t.Fatalf("counters removed event = %+v, want three +1/+1 counters from the source", removed)
	}
	wantPaidCounters(t, original, 3)

	source.Counters.Add(counter.PlusOnePlusOne, 2)
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(original.ID)})
	e.resolveTopOfStack(g, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || copied.Kind != game.StackActivatedAbility || copied.ID == original.ID {
		t.Fatalf("copied ability = %+v, want an independent activated-ability copy", copied)
	}
	wantPaidCounters(t, copied, 3)

	clone := g.Clone()
	e.resolveTopOfStack(clone, &TurnLog{})
	if got := clone.Players[game.Player2].Life; got != life-3 {
		t.Fatalf("cloned copy dealt %d damage, want 3", life-got)
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != life-3 {
		t.Fatalf("copy dealt %d damage, want the original's paid count 3", life-got)
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != life-6 {
		t.Fatalf("original dealt %d damage, want 3 despite the source's new counters", life-got-3)
	}
	if source.Counters.Get(counter.PlusOnePlusOne) != 2 {
		t.Fatalf("source +1/+1 counters = %d, want the two added after payment", source.Counters.Get(counter.PlusOnePlusOne))
	}
}

// TestRemoveAllCountersCostPayableWithZero proves the cost is paid with no
// counters of its kind: activation succeeds, records zero and deals no damage.
func TestRemoveAllCountersCostPayableWithZero(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, removeAllCounterProbe(0))
	source.Counters.Add(counter.Charge, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	life := g.Players[game.Player2].Life

	offered := false
	for _, act := range e.legalActions(g, game.Player1) {
		if payload, ok := act.ActivateAbilityPayload(); ok && payload.SourceID == source.ObjectID {
			offered = true
		}
	}
	if !offered {
		t.Fatal("legal actions omit the zero-counter remove-all activation")
	}
	obj := activateRemoveAllProbe(t, g, e, source)
	wantPaidCounters(t, obj, 0)
	if !source.Tapped || source.Counters.Get(counter.Charge) != 1 {
		t.Fatalf("source tapped=%v counters=%v, want tapped with its charge counter kept", source.Tapped, source.Counters)
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != life {
		t.Fatalf("zero-counter activation dealt %d damage, want 0", life-got)
	}
}

// TestRemoveAllCountersCostUnaffordableLeavesCounters proves an activation
// whose mana part can't be paid is rejected atomically (CR 118.3, 601.2h).
func TestRemoveAllCountersCostUnaffordableLeavesCounters(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, removeAllCounterProbe(0))
	source.Counters.Add(counter.PlusOnePlusOne, 3)

	if e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.PlayerTarget(game.Player2)}, 0)) {
		t.Fatal("activation without {R} was accepted")
	}
	if source.Tapped || source.Counters.Get(counter.PlusOnePlusOne) != 3 || g.Stack.Size() != 0 {
		t.Fatalf("source tapped=%v counters=%v stack=%d, want the failed activation to change nothing",
			source.Tapped, source.Counters, g.Stack.Size())
	}
}

// TestRemoveAllCountersCostCountsEachFive applies Sage of Hours' "for each five
// counters removed this way" rounding to the paid count: twelve counters give
// two and three give none (Sage of Hours ruling 2014-04-26, CR 107.1a).
func TestRemoveAllCountersCostCountsEachFive(t *testing.T) {
	for _, tc := range []struct{ counters, want int }{{12, 2}, {3, 0}} {
		g := mainPhaseGame(game.Player1)
		setMainPhasePriority(g, game.Player1)
		e := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, removeAllCounterProbe(5))
		source.Counters.Add(counter.PlusOnePlusOne, tc.counters)
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		life := g.Players[game.Player2].Life

		activateRemoveAllProbe(t, g, e, source)
		e.resolveTopOfStack(g, &TurnLog{})
		if got := life - g.Players[game.Player2].Life; got != tc.want || source.Counters.Get(counter.PlusOnePlusOne) != 0 {
			t.Fatalf("%d counters: dealt %d with %d left, want %d with none left", tc.counters, got, source.Counters.Get(counter.PlusOnePlusOne), tc.want)
		}
	}
}
