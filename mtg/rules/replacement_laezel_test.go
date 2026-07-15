package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// laezelCounterReplacementCardDef models Lae'zel, Vlaakith's Champion's counter
// replacement: "If you would put one or more counters on a creature or
// planeswalker you control or on yourself, put that many plus one of each of
// those kinds of counters on that permanent or player instead."
func laezelCounterReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Lae'zel, Vlaakith's Champion",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.ControlledPermanentTypesOrControllerCounterPlacementReplacement(
				"If you would put one or more counters on a creature or planeswalker you control or on yourself, put that many plus one of each of those kinds of counters on that permanent or player instead.",
				0,
				1,
				[]types.Card{types.Creature, types.Planeswalker},
				game.TriggerControllerYou,
			),
		},
	}}
}

// TestLaezelAddsToControlledCreatureAndPlaneswalker proves the recipient union
// covers both a controlled creature and a controlled planeswalker, adding one of
// the placed counter kind beyond the original amount.
func TestLaezelAddsToControlledCreatureAndPlaneswalker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, laezelCounterReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "My Creature",
		Types: []types.Card{types.Creature},
	}})
	planeswalker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "My Planeswalker",
		Types: []types.Card{types.Planeswalker},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 2) {
		t.Fatal("addCountersToPermanent(creature) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters on controlled creature = %d, want 3", got)
	}
	if !addCountersToPermanent(g, planeswalker, counter.Loyalty, 3) {
		t.Fatal("addCountersToPermanent(planeswalker) = false, want true")
	}
	if got := planeswalker.Counters.Get(counter.Loyalty); got != 4 {
		t.Fatalf("loyalty counters on controlled planeswalker = %d, want 4", got)
	}
}

// TestLaezelAddsToControllerPlayerCounters proves the "or on yourself" arm adds
// one to poison, energy, and experience counters placed on the controller.
func TestLaezelAddsToControllerPlayerCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, laezelCounterReplacementCardDef())
	controller := g.Players[game.Player1]

	for _, tc := range []struct {
		name string
		kind counter.Kind
		get  func() int
	}{
		{"poison", counter.Poison, func() int { return controller.PoisonCounters }},
		{"energy", counter.Energy, func() int { return controller.EnergyCounters }},
		{"experience", counter.Experience, func() int { return controller.ExperienceCounters }},
	} {
		if !addCountersToPlayer(g, controller, tc.kind, 1) {
			t.Fatalf("addCountersToPlayer(%s) = false, want true", tc.name)
		}
		if got := tc.get(); got != 2 {
			t.Fatalf("%s counters on controller = %d, want 2", tc.name, got)
		}
	}
}

// TestLaezelPreservesEveryCounterKindInMultiKindPlacement proves a placement of
// several distinct counter kinds preserves each kind and adds one of each.
func TestLaezelPreservesEveryCounterKindInMultiKindPlacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, laezelCounterReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Multi-Kind Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 2) {
		t.Fatal("addCountersToPermanent(+1/+1) = false, want true")
	}
	if !addCountersToPermanent(g, creature, counter.Charge, 1) {
		t.Fatal("addCountersToPermanent(charge) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}
	if got := creature.Counters.Get(counter.Charge); got != 2 {
		t.Fatalf("charge counters = %d, want 2", got)
	}
}

// TestLaezelDoesNotAffectOpponentRecipients proves neither an opponent's
// creature nor the opponent as a player benefits from the replacement.
func TestLaezelDoesNotAffectOpponentRecipients(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, laezelCounterReplacementCardDef())
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, opponentCreature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(opponent creature) = false, want true")
	}
	if got := opponentCreature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters on opponent creature = %d, want 1 (not modified)", got)
	}
	if !addCountersToPlayer(g, g.Players[game.Player2], counter.Poison, 1) {
		t.Fatal("addCountersToPlayer(opponent) = false, want true")
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 1 {
		t.Fatalf("poison counters on opponent = %d, want 1 (not modified)", got)
	}
}

// TestLaezelFollowsLiveSourceController proves the recipient's "you control"
// and "yourself" scopes track the source's live controller after a control
// change, so the replacement then benefits the new controller's permanents.
func TestLaezelFollowsLiveSourceController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, laezelCounterReplacementCardDef())
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player2),
		Duration:         game.DurationPermanent,
	})
	oldControllerCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Old Creature",
		Types: []types.Card{types.Creature},
	}})
	newControllerCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "New Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, oldControllerCreature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(old controller) = false, want true")
	}
	if got := oldControllerCreature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("old controller +1/+1 counters = %d, want 1", got)
	}
	if !addCountersToPermanent(g, newControllerCreature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(new controller) = false, want true")
	}
	if got := newControllerCreature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("new controller +1/+1 counters = %d, want 2", got)
	}
	if !addCountersToPlayer(g, g.Players[game.Player2], counter.Poison, 1) {
		t.Fatal("addCountersToPlayer(new controller) = false, want true")
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("new controller poison counters = %d, want 2", got)
	}
}

// TestLaezelStacksWithAnotherApplicableReplacement proves multiple applicable
// counter-placement replacements compose under CR 616.1 ordering, applying once
// each and recording the controller's ordering decision.
func TestLaezelStacksWithAnotherApplicableReplacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, laezelCounterReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, laezelCounterReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent() = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters from two stacked replacements = %d, want 3", got)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player1 {
		t.Fatalf("replacement decision player = %v, want Player1", got)
	}
}
