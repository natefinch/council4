package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// randomModalSagaCard models the shape cardgen lowers for Final Fantasy "Summon"
// sagas whose chapters read "Choose one at random —": a single chapter ability
// whose AbilityContent is modal with RandomModes set. Each mode here targets a
// creature and places a distinct counter so a test can identify which mode the
// game's random source selected.
func randomModalSagaCard() *game.CardDef {
	creatureTarget := []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	mode := func(text string, kind counter.Kind, amount int) game.Mode {
		return game.Mode{
			Text:    text,
			Targets: creatureTarget,
			Sequence: []game.Instruction{{Primitive: game.AddCounter{
				Object:      game.TargetPermanentReference(0),
				Amount:      game.Fixed(amount),
				CounterKind: kind,
			}}},
		}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Random Saga",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Saga},
		ChapterAbilities: []game.ChapterAbility{{
			Text:     "Choose one at random.",
			Chapters: []int{1, 2, 3},
			Content: game.AbilityContent{
				MinModes:    1,
				MaxModes:    1,
				RandomModes: true,
				Modes: []game.Mode{
					mode("Three +1/+1 counters.", counter.PlusOnePlusOne, 3),
					mode("A shield counter.", counter.Shield, 1),
					mode("A stun counter.", counter.Stun, 1),
				},
			},
		}},
	}}
}

// chosenModeCounter resolves the random-modal chapter once with the supplied
// random source and returns the counter kind whose mode resolved, asserting
// exactly one of the three modes applied its effect to the target creature.
func chosenModeCounter(t *testing.T, rng *rand.Rand) counter.Kind {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rng)
	saga := addCombatPermanent(g, game.Player1, randomModalSagaCard())
	creature := addCreaturePermanent(g, game.Player1)

	if !addCountersToPermanent(g, saga, counter.Lore, 1) {
		t.Fatal("failed to add lore counter to saga")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("random-modal chapter did not trigger on the lore counter")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	kinds := []counter.Kind{counter.PlusOnePlusOne, counter.Shield, counter.Stun}
	var applied counter.Kind
	hits := 0
	for _, kind := range kinds {
		if creature.Counters.Get(kind) > 0 {
			applied = kind
			hits++
		}
	}
	if hits != 1 {
		t.Fatalf("expected exactly one mode to apply a counter, got %d", hits)
	}
	return applied
}

// TestRandomModalChapterSelectsAndAppliesOneMode proves a saga chapter whose
// content has RandomModes set triggers, selects a single mode through the
// engine's random source, segments that mode's target, and applies the effect
// without panicking.
func TestRandomModalChapterSelectsAndAppliesOneMode(t *testing.T) {
	_ = chosenModeCounter(t, rand.New(rand.NewPCG(1, 2)))
}

// TestRandomModalChapterReachesEveryMode proves the random selection is not
// pinned to one mode: across several seeds every mode is selected at least
// once, so the at-random primitive draws from the full set of modes.
func TestRandomModalChapterReachesEveryMode(t *testing.T) {
	seen := map[counter.Kind]bool{}
	for seed := range uint64(40) {
		seen[chosenModeCounter(t, rand.New(rand.NewPCG(seed, seed+1)))] = true
	}
	for _, kind := range []counter.Kind{counter.PlusOnePlusOne, counter.Shield, counter.Stun} {
		if !seen[kind] {
			t.Fatalf("random-modal selection never chose the %v mode across seeds", kind)
		}
	}
}
