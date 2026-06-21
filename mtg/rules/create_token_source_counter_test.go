package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestCreateTokenSourceCounterCountReadsLastKnownCounters verifies the
// "When ~ dies, create X 1/1 ... tokens, where X is the number of +1/+1
// counters on it." family (Chasm Skulker): the self-death trigger creates a
// number of tokens equal to the dying permanent's +1/+1 counters, read from
// last-known information once it has left the battlefield.
func TestCreateTokenSourceCounterCountReadsLastKnownCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	squid := &game.CardDef{CardFace: game.CardFace{
		Name:     "Squid",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Squid},
	}}
	skulker := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		Source:                game.TriggerSourceSelf,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:        game.DynamicAmountObjectCounters,
			Multiplier:  1,
			CounterKind: counter.PlusOnePlusOne,
			Object:      game.EventPermanentReference(),
		}),
		Source: game.TokenDef(squid),
	}}}, nil)

	if !addCountersToPermanent(g, skulker, counter.PlusOnePlusOne, 3) {
		t.Fatal("addCountersToPermanent(+1/+1) = false, want true")
	}

	destroyPermanent(g, skulker.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-death token trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokenPermanentsNamed(g, "Squid"); got != 3 {
		t.Fatalf("Squid tokens = %d, want 3 (equal to +1/+1 counters at death)", got)
	}
}

// TestCreateTokenSourceCounterCountScalesWithCounters verifies the token count
// scales with the dying permanent's +1/+1 counters for a larger count.
func TestCreateTokenSourceCounterCountScalesWithCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	squid := &game.CardDef{CardFace: game.CardFace{
		Name:     "Squid",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Squid},
	}}
	skulker := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		Source:                game.TriggerSourceSelf,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:        game.DynamicAmountObjectCounters,
			Multiplier:  1,
			CounterKind: counter.PlusOnePlusOne,
			Object:      game.EventPermanentReference(),
		}),
		Source: game.TokenDef(squid),
	}}}, nil)

	if !addCountersToPermanent(g, skulker, counter.PlusOnePlusOne, 5) {
		t.Fatal("addCountersToPermanent(+1/+1) = false, want true")
	}

	destroyPermanent(g, skulker.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-death token trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokenPermanentsNamed(g, "Squid"); got != 5 {
		t.Fatalf("Squid tokens = %d, want 5 (equal to +1/+1 counters at death)", got)
	}
}
