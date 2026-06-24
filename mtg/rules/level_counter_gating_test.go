package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestLevelCounterGatedAbilityActiveOnlyInBand verifies the gating semantics the
// Level Up slice produces (CR 711.4): a band ability gated with
// SourceLevelCountersAtLeast/LessThan is active only while the source's level
// counter count falls inside the band, and the level-up ability adds one level
// counter per resolution.
func TestLevelCounterGatedAbilityActiveOnlyInBand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Leveler",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Merfolk}},
	})
	card := g.CardInstances[cardID]
	source, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed")
	}

	lowBand := opt.Val(game.Condition{SourceLevelCountersAtLeast: 1, SourceLevelCountersLessThan: 3})
	highBand := opt.Val(game.Condition{SourceLevelCountersAtLeast: 3})

	if activationConditionSatisfied(g, game.Player1, source, lowBand) {
		t.Fatal("LEVEL 1-2 band ability should be inactive with no level counters")
	}
	if activationConditionSatisfied(g, game.Player1, source, highBand) {
		t.Fatal("LEVEL 3+ band ability should be inactive with no level counters")
	}

	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	addLevel := game.AddCounter{Object: game.SourcePermanentReference(), Amount: game.Fixed(1), CounterKind: counter.Level}
	resolveInstruction(engine, g, obj, addLevel, &TurnLog{})

	if got := source.Counters.Get(counter.Level); got != 1 {
		t.Fatalf("level counters = %d, want 1", got)
	}
	if !activationConditionSatisfied(g, game.Player1, source, lowBand) {
		t.Fatal("LEVEL 1-2 band ability should be active at level 1")
	}
	if activationConditionSatisfied(g, game.Player1, source, highBand) {
		t.Fatal("LEVEL 3+ band ability should be inactive at level 1")
	}

	resolveInstruction(engine, g, obj, addLevel, &TurnLog{})
	resolveInstruction(engine, g, obj, addLevel, &TurnLog{})

	if got := source.Counters.Get(counter.Level); got != 3 {
		t.Fatalf("level counters = %d, want 3", got)
	}
	if activationConditionSatisfied(g, game.Player1, source, lowBand) {
		t.Fatal("LEVEL 1-2 band ability should be inactive at level 3")
	}
	if !activationConditionSatisfied(g, game.Player1, source, highBand) {
		t.Fatal("LEVEL 3+ band ability should be active at level 3")
	}
}
