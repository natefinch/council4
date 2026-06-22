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

// multikickerCreature is a {1} creature with Multikicker {1} whose
// enters-with-counters replacement places a +1/+1 counter for each time it was
// kicked (the Gnarlid Pack / Everflowing Chalice payoff shape).
func multikickerCreature() *game.CardDef {
	def := creatureSpellDef("Multikick Beast", types.Beast)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	def.StaticAbilities = []game.StaticAbility{{
		KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1)}, Multi: true}},
	}}
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersReplacement(
			"This creature enters with a +1/+1 counter on it for each time it was kicked.",
			game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
				Kind:       game.DynamicAmountTimesKicked,
				Multiplier: 1,
			})},
		),
	}
	return def
}

// castMultikickerCreature casts the Multikicker creature paying its {1} base
// cost plus its {1} kicker kickCount times, then resolves it.
func castMultikickerCreature(t *testing.T, kickCount int) *game.Permanent {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1+kickCount)
	spellID := addCardToHand(g, game.Player1, multikickerCreature())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastMultikickedSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, nil, 0, nil, kickCount)) {
		t.Fatalf("applyAction(cast multikicked %d) = false, want true", kickCount)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("Multikicker creature did not resolve to the battlefield")
	}
	return permanent
}

// TestMultikickerEntersWithCounterPerKick verifies the kick count scales the
// "for each time it was kicked" enters-with-counters amount: paying the kicker
// N times places N +1/+1 counters.
func TestMultikickerEntersWithCounterPerKick(t *testing.T) {
	t.Parallel()
	for _, kicks := range []int{1, 2, 3} {
		permanent := castMultikickerCreature(t, kicks)
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != kicks {
			t.Fatalf("counters = %d, want %d for kicking %d times", got, kicks, kicks)
		}
	}
}

// TestMultikickerUnkickedEntersWithoutCounters verifies a Multikicker spell cast
// without paying the kicker enters with no counters (kick count zero).
func TestMultikickerUnkickedEntersWithoutCounters(t *testing.T) {
	t.Parallel()
	permanent := castMultikickerCreature(t, 0)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("counters = %d, want 0 for an unkicked cast", got)
	}
}
