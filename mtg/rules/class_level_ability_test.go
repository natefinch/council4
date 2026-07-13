package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// newClassGatedDoubler puts a Class enchantment carrying a level-3-gated
// "double any counters you would put" replacement onto the battlefield for
// Player1 and returns its permanent, mirroring the executable lowering of
// Innkeeper's Talent's level-3 band.
func newClassGatedDoubler(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Class",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Class},
		ReplacementAbilities: []game.ReplacementAbility{
			game.ClassLevelGatedReplacement(
				game.AnyCounterPlacementReplacement("double counters", 2, 0, game.TriggerControllerYou),
				3,
			),
		},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	permanent, ok := createCardPermanent(g, g.CardInstances[cardID], game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed")
	}
	return permanent
}

// TestClassWardGrantCoversNonCreaturePermanentsWithCounters proves the level-2
// band "Permanents you control with counters on them have ward {1}" grants ward
// to every permanent type (not just creatures) that currently bears a counter,
// updates continuously as counters appear and disappear, and is scoped to the
// controller.
func TestClassWardGrantCoversNonCreaturePermanentsWithCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	source := newClassPermanent(t, g)

	artifactID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Trinket",
		Types: []types.Card{types.Artifact},
	}})
	artifact, ok := createCardPermanent(g, g.CardInstances[artifactID], game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed for artifact")
	}

	plainID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bauble",
		Types: []types.Card{types.Artifact},
	}})
	plain, ok := createCardPermanent(g, g.CardInstances[plainID], game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed for second artifact")
	}

	opponentID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Rival Relic",
		Types: []types.Card{types.Artifact},
	}})
	opponent, ok := createCardPermanent(g, g.CardInstances[opponentID], game.Player2, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed for opponent artifact")
	}
	opponent.Counters.Add(counter.Charge, 1)

	ward := game.WardStaticAbility(cost.Mana{cost.O(1)})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: source.ObjectID,
		Layer:          game.LayerAbility,
		Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{MatchAnyCounter: true}),
		AddAbilities:   []game.Ability{&ward},
	})

	if permanentHasGrantedWard(g, artifact) {
		t.Fatal("artifact without a counter should not have ward")
	}

	artifact.Counters.Add(counter.Charge, 1)
	if !permanentHasGrantedWard(g, artifact) {
		t.Fatal("non-creature permanent with a counter should gain ward")
	}
	if permanentHasGrantedWard(g, plain) {
		t.Fatal("controlled permanent without a counter should not have ward")
	}
	if permanentHasGrantedWard(g, opponent) {
		t.Fatal("opponent's counter-bearing permanent is outside the controller scope")
	}

	artifact.Counters.Remove(counter.Charge, 1)
	if permanentHasGrantedWard(g, artifact) {
		t.Fatal("ward should drop when the last counter leaves")
	}
}

// TestClassLevelGatedCounterDoublingActivatesAtLevelThree proves the level-3
// counter-doubling replacement is inactive until the source Class reaches level
// 3, then doubles counters placed on a permanent the controller controls.
func TestClassLevelGatedCounterDoublingActivatesAtLevelThree(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassGatedDoubler(t, g)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	addCountersToPermanent(g, target, counter.PlusOnePlusOne, 1)
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("at level 1: counters = %d, want 1 (doubling inactive)", got)
	}

	raiseClassLevel(t, engine, g, source, 2)
	addCountersToPermanent(g, target, counter.PlusOnePlusOne, 1)
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("at level 2: counters = %d, want 2 (doubling still inactive)", got)
	}

	raiseClassLevel(t, engine, g, source, 3)
	addCountersToPermanent(g, target, counter.PlusOnePlusOne, 2)
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 6 {
		t.Fatalf("at level 3: counters = %d, want 6 (2 existing + doubled 2->4)", got)
	}
}

// TestClassLevelGatedCounterDoublingAppliesToPlayers proves the level-3
// replacement doubles counters the controller would put on a player, not only on
// permanents (CR 614.16).
func TestClassLevelGatedCounterDoublingAppliesToPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassGatedDoubler(t, g)
	raiseClassLevel(t, engine, g, source, 3)

	opponent, ok := playerByID(g, game.Player2)
	if !ok {
		t.Fatal("player lookup failed")
	}
	// Player1 (the Class controller, "you") puts poison counters on the opponent.
	addCountersToPlayerControlledBy(g, game.Player1, opponent, counter.Poison, 3)
	if got := opponent.PoisonCounters; got != 6 {
		t.Fatalf("poison counters on player = %d, want 6 (3 doubled)", got)
	}
}

// TestClassLevelGatedCounterDoublingStopsWhenSourceLeaves proves the replacement
// disappears when the source Class leaves the battlefield: the gate resolves
// from the source permanent, which is no longer present.
func TestClassLevelGatedCounterDoublingStopsWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassGatedDoubler(t, g)
	raiseClassLevel(t, engine, g, source, 3)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("moving the Class to the graveyard failed")
	}

	addCountersToPermanent(g, target, counter.PlusOnePlusOne, 1)
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("after source left: counters = %d, want 1 (doubling gone)", got)
	}
}

// TestClassLevelGatedCounterDoublingFollowsControllerScope proves the "you"
// controller scope tracks the source's current controller: after a control
// change the doubling applies to the new controller's counter placements and no
// longer to the former controller's.
func TestClassLevelGatedCounterDoublingFollowsControllerScope(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassGatedDoubler(t, g)
	raiseClassLevel(t, engine, g, source, 3)
	source.Controller = game.Player2

	ownTarget := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCountersToPermanent(g, ownTarget, counter.PlusOnePlusOne, 1)
	if got := ownTarget.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("former controller's placement doubled = %d, want 1 (scope moved)", got)
	}

	newTarget := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCountersToPermanent(g, newTarget, counter.PlusOnePlusOne, 1)
	if got := newTarget.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("new controller's placement doubled = %d, want 2", got)
	}
}

// TestClassLevelGatedCounterDoublingStacksWithAnotherDoubler proves the
// class-gated replacement composes with an ordinary counter doubler (Doubling
// Season) already in the engine: at level 3 both apply and the count quadruples.
func TestClassLevelGatedCounterDoublingStacksWithAnotherDoubler(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassGatedDoubler(t, g)
	raiseClassLevel(t, engine, g, source, 3)

	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID:                g.IDGen.Next(),
		Description:       "Doubling Season",
		MatchEvent:        game.EventCountersAdded,
		ControllerFilter:  game.TriggerControllerYou,
		Controller:        game.Player1,
		CounterMultiplier: 2,
		Duration:          game.DurationPermanent,
	})

	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCountersToPermanent(g, target, counter.PlusOnePlusOne, 1)
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("stacked doublers: counters = %d, want 4 (1 doubled twice)", got)
	}
}
