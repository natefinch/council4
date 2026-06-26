package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// adamantColorCreature is a {2}{W} creature whose Adamant replacement places a
// +1/+1 counter only if at least three white mana was spent to cast it
// (Ardenvale Paladin's headline ability, modeled with a cheaper cost).
func adamantColorCreature() *game.CardDef {
	def := creatureSpellDef("Adamant Paladin", types.Knight)
	def.ManaCost = opt.Val(cost.Mana{cost.O(2), cost.W})
	def.Power = opt.Val(game.PT{Value: 2})
	def.Toughness = opt.Val(game.PT{Value: 2})
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersIfReplacement(
			"Adamant — If at least three white mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.",
			&game.Condition{SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.White, Count: 3}},
			game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1},
		),
	}
	return def
}

// adamantSameColorCreature is a {3} creature whose Adamant replacement places a
// +1/+1 counter only if at least three mana of the same color was spent to cast
// it (Henge Walker's headline ability).
func adamantSameColorCreature() *game.CardDef {
	def := creatureSpellDef("Adamant Golem", types.Golem)
	def.ManaCost = opt.Val(cost.Mana{cost.O(3)})
	def.Power = opt.Val(game.PT{Value: 2})
	def.Toughness = opt.Val(game.PT{Value: 2})
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersIfReplacement(
			"Adamant — If at least three mana of the same color was spent to cast this spell, this creature enters with a +1/+1 counter on it.",
			&game.Condition{SpellSameColorManaSpentAtLeast: 3},
			game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1},
		),
	}
	return def
}

// castAdamantCreatureWithPool casts the supplied Adamant creature, paying its
// cost from the supplied pool, and resolves it to the battlefield.
func castAdamantCreatureWithPool(t *testing.T, def *game.CardDef, add func(pool *mana.Pool)) *game.Permanent {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	add(&g.Players[game.Player1].ManaPool)
	spellID := addCardToHand(g, game.Player1, def)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast Adamant creature) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("Adamant creature did not resolve to the battlefield")
	}
	return permanent
}

// TestAdamantColorMetEntersWithCounter verifies that paying three white mana on
// a {2}{W} Adamant creature satisfies the condition and adds a +1/+1 counter.
func TestAdamantColorMetEntersWithCounter(t *testing.T) {
	t.Parallel()
	permanent := castAdamantCreatureWithPool(t, adamantColorCreature(), func(pool *mana.Pool) {
		pool.Add(mana.W, 3)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("counters = %d, want 1 for three white mana spent", got)
	}
}

// TestAdamantColorUnmetNoCounter verifies that paying only one white mana (with
// the generic cost paid by colorless) fails the Adamant condition.
func TestAdamantColorUnmetNoCounter(t *testing.T) {
	t.Parallel()
	permanent := castAdamantCreatureWithPool(t, adamantColorCreature(), func(pool *mana.Pool) {
		pool.Add(mana.W, 1)
		pool.Add(mana.C, 2)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("counters = %d, want 0 for only one white mana spent", got)
	}
}

// TestAdamantColorWrongColorNoCounter verifies that three mana of a different
// color does not satisfy a white Adamant condition, even though the white pip is
// paid with white.
func TestAdamantColorWrongColorNoCounter(t *testing.T) {
	t.Parallel()
	permanent := castAdamantCreatureWithPool(t, adamantColorCreature(), func(pool *mana.Pool) {
		pool.Add(mana.W, 1)
		pool.Add(mana.U, 2)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("counters = %d, want 0 for one white and two blue mana spent", got)
	}
}

// TestAdamantSameColorMetEntersWithCounter verifies that paying three mana of a
// single color satisfies the same-color Adamant condition.
func TestAdamantSameColorMetEntersWithCounter(t *testing.T) {
	t.Parallel()
	permanent := castAdamantCreatureWithPool(t, adamantSameColorCreature(), func(pool *mana.Pool) {
		pool.Add(mana.R, 3)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("counters = %d, want 1 for three red mana spent", got)
	}
}

// TestAdamantSameColorMixedNoCounter verifies that three mana spread across
// different colors fails the same-color Adamant condition.
func TestAdamantSameColorMixedNoCounter(t *testing.T) {
	t.Parallel()
	permanent := castAdamantCreatureWithPool(t, adamantSameColorCreature(), func(pool *mana.Pool) {
		pool.Add(mana.W, 1)
		pool.Add(mana.U, 1)
		pool.Add(mana.B, 1)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("counters = %d, want 0 for three different colors spent", got)
	}
}
