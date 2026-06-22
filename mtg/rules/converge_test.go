package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// convergeCreature is a {3} creature whose Converge enters-with-counters
// replacement places a +1/+1 counter for each color of mana spent to cast it
// (Crystalline Crawler's headline ability).
func convergeCreature() *game.CardDef {
	def := creatureSpellDef("Converge Crawler", types.Construct)
	def.ManaCost = opt.Val(cost.Mana{cost.O(3)})
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersReplacement(
			"This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.",
			game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
				Kind:       game.DynamicAmountColorsOfManaSpentToCast,
				Multiplier: 1,
			})},
		),
	}
	return def
}

// castConvergeCreatureWithPool casts the Converge creature paying its {3} cost
// from the supplied pool and resolves it, returning the resulting permanent.
func castConvergeCreatureWithPool(t *testing.T, add func(pool *mana.Pool)) *game.Permanent {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	add(&g.Players[game.Player1].ManaPool)
	spellID := addCardToHand(g, game.Player1, convergeCreature())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast Converge creature) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("Converge creature did not resolve to the battlefield")
	}
	return permanent
}

// TestConvergeCountsDistinctColorsSpent verifies the Converge count places one
// +1/+1 counter per distinct color of mana spent on the {3} cost, including when
// colored mana pays a generic cost.
func TestConvergeCountsDistinctColorsSpent(t *testing.T) {
	t.Parallel()
	permanent := castConvergeCreatureWithPool(t, func(pool *mana.Pool) {
		pool.Add(mana.W, 1)
		pool.Add(mana.U, 1)
		pool.Add(mana.B, 1)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("counters = %d, want 3 for three colors spent", got)
	}
}

// TestConvergeIgnoresColorlessMana verifies colorless mana contributes no color,
// so a creature paid entirely with colorless mana enters with no counters.
func TestConvergeIgnoresColorlessMana(t *testing.T) {
	t.Parallel()
	permanent := castConvergeCreatureWithPool(t, func(pool *mana.Pool) {
		pool.Add(mana.C, 3)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("counters = %d, want 0 for only colorless mana spent", got)
	}
}

// TestConvergeCountsRepeatedColorOnce verifies repeated mana of the same color
// counts once: two white and one blue is two distinct colors.
func TestConvergeCountsRepeatedColorOnce(t *testing.T) {
	t.Parallel()
	permanent := castConvergeCreatureWithPool(t, func(pool *mana.Pool) {
		pool.Add(mana.W, 2)
		pool.Add(mana.U, 1)
	})
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("counters = %d, want 2 for two distinct colors spent", got)
	}
}
