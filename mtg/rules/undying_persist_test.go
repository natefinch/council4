package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func undyingTestCreature(body *game.TriggeredAbility) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{
		Name:               "Test Undying Beast",
		Types:              []types.Card{types.Creature},
		Subtypes:           []types.Sub{types.Beast},
		Power:              opt.Val(pt),
		Toughness:          opt.Val(pt),
		TriggeredAbilities: []game.TriggeredAbility{*body},
	}}
}

// TestUndyingReturnsWithCounterWhenItHadNone proves the canonical undying body
// (CR 702.92): a creature with no +1/+1 counters that dies returns to the
// battlefield with one +1/+1 counter.
func TestUndyingReturnsWithCounterWhenItHadNone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player1, undyingTestCreature(&game.UndyingTriggeredBody))
	cardID := creature.CardInstanceID
	creature.MarkedDamage = 1

	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("creature with lethal damage remained on the battlefield")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("undying did not trigger when the creature died without a +1/+1 counter")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentForCard(g, cardID)
	if returned == nil {
		t.Fatal("undying did not return the creature to the battlefield")
	}
	if got := returned.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("returned creature has %d +1/+1 counters, want 1", got)
	}
	if !hasKeyword(g, returned, game.Undying) {
		t.Fatal("returned creature does not report the Undying keyword")
	}
}

// TestUndyingDoesNotTriggerWhenItHadCounter proves the undying intervening-if:
// a creature that already had a +1/+1 counter when it died does not return.
func TestUndyingDoesNotTriggerWhenItHadCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player1, undyingTestCreature(&game.UndyingTriggeredBody))
	cardID := creature.CardInstanceID
	addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1)
	creature.MarkedDamage = 2

	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("creature with lethal damage remained on the battlefield")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("undying triggered even though the creature died with a +1/+1 counter")
	}
	if permanentForCard(g, cardID) != nil {
		t.Fatal("creature returned to the battlefield despite having had a +1/+1 counter")
	}
}

// TestPersistReturnsWithMinusCounterWhenItHadNone proves the canonical persist
// body (CR 702.78): a creature with no -1/-1 counters that dies returns with one
// -1/-1 counter.
func TestPersistReturnsWithMinusCounterWhenItHadNone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player1, undyingTestCreature(&game.PersistTriggeredBody))
	cardID := creature.CardInstanceID
	creature.MarkedDamage = 1

	engine.applyStateBasedActions(g)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("persist did not trigger when the creature died without a -1/-1 counter")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentForCard(g, cardID)
	if returned == nil {
		t.Fatal("persist did not return the creature to the battlefield")
	}
	if got := returned.Counters.Get(counter.MinusOneMinusOne); got != 1 {
		t.Fatalf("returned creature has %d -1/-1 counters, want 1", got)
	}
	if !hasKeyword(g, returned, game.Persist) {
		t.Fatal("returned creature does not report the Persist keyword")
	}
}
