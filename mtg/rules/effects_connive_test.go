package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestConniveDiscardingNonlandAddsCounter proves the connive keyword action
// (CR 702.154) draws a card, discards a card, and places a +1/+1 counter on the
// conniving permanent because the discarded card was a nonland.
func TestConniveDiscardingNonlandAddsCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Conniver",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Lightning Bolt",
		Types: []types.Card{types.Instant},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.Connive{
		Object: game.SourcePermanentReference(),
		Player: game.ControllerReference(),
		Amount: game.Fixed(1),
	}, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(drawn) {
		t.Fatal("connive did not discard the drawn nonland card")
	}
	if g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("discarded card still in hand")
	}
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1 after discarding a nonland card", got)
	}
}

// TestConniveDiscardingLandAddsNoCounter proves connive places no counter when
// the discarded card is a land (CR 702.154).
func TestConniveDiscardingLandAddsNoCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Conniver",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Island",
		Types: []types.Card{types.Land},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.Connive{
		Object: game.SourcePermanentReference(),
		Player: game.ControllerReference(),
		Amount: game.Fixed(1),
	}, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(drawn) {
		t.Fatal("connive did not discard the drawn land card")
	}
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters = %d, want 0 after discarding a land card", got)
	}
}
