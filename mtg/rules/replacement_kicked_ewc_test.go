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

// kickedEntersWithCountersCreature is a {1} creature with Kicker {1} whose
// enters-with-counters replacement places two +1/+1 counters only when the
// spell was kicked (the Gnarlid Colony / Kavu Aggressor shape).
func kickedEntersWithCountersCreature() *game.CardDef {
	def := creatureSpellDef("Kicked EWC Beast", types.Beast)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	def.StaticAbilities = []game.StaticAbility{{
		KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1)}}},
	}}
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersIfReplacement(
			"If this creature was kicked, it enters with two +1/+1 counters on it.",
			&game.Condition{EventPermanentWasKicked: true},
			game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2},
		),
	}
	return def
}

// castKickedEntersWithCountersCreature casts the creature paying its {1} base
// cost, optionally paying the {1} kicker, then resolves it.
func castKickedEntersWithCountersCreature(t *testing.T, kicked bool) *game.Permanent {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	amount := 1
	if kicked {
		amount = 2
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, amount)
	spellID := addCardToHand(g, game.Player1, kickedEntersWithCountersCreature())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	castAction := action.CastSpell(spellID, nil, 0, nil)
	if kicked {
		castAction = action.CastKickedSpell(spellID, nil, 0, nil)
	}
	if !engine.applyAction(g, game.Player1, castAction) {
		t.Fatalf("applyAction(cast kicked=%v) = false, want true", kicked)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("kicked EWC creature did not resolve to the battlefield")
	}
	return permanent
}

// TestKickedEntersWithCountersAppliesWhenKicked verifies the conditional
// enters-with-counters replacement places its counters when the spell was
// kicked.
func TestKickedEntersWithCountersAppliesWhenKicked(t *testing.T) {
	t.Parallel()
	permanent := castKickedEntersWithCountersCreature(t, true)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("counters = %d, want 2 for a kicked cast", got)
	}
}

// TestKickedEntersWithCountersSkippedWhenUnkicked verifies the conditional
// enters-with-counters replacement places no counters when the spell was cast
// without paying the kicker.
func TestKickedEntersWithCountersSkippedWhenUnkicked(t *testing.T) {
	t.Parallel()
	permanent := castKickedEntersWithCountersCreature(t, false)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("counters = %d, want 0 for an unkicked cast", got)
	}
}
