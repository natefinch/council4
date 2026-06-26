package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestCastForFreeTargetedOpponentGraveyard verifies the targeted free cast of an
// instant in an opponent's graveyard (Memory Plunder): resolving the source
// moves the targeted card from the opponent's graveyard onto the stack under the
// casting player's control.
func TestCastForFreeTargetedOpponentGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	targetID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Bolt",
		Types: []types.Card{types.Instant},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.CastForFree{
		Player: game.ControllerReference(),
		Card:   game.CardReference{Kind: game.CardReferenceTarget},
		Zone:   zone.Graveyard,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Instant}, Controller: game.ControllerOpponent}),
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("targeted card still in opponent's graveyard")
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("no spell on the stack after the free cast")
	}
	if top.SourceID != targetID {
		t.Fatalf("stack top SourceID = %v, want the cast graveyard card %v", top.SourceID, targetID)
	}
	if top.Controller != game.Player1 {
		t.Fatalf("cast spell controller = %v, want the casting player %v", top.Controller, game.Player1)
	}
}

// TestCastForFreeTargetedExileOnResolution verifies the "If that spell would be
// put into your graveyard, exile it instead." rider (Torrential Gearhulk): the
// free-cast spell carries ExileOnResolution, so after it resolves the instant
// moves to exile rather than its owner's graveyard.
func TestCastForFreeTargetedExileOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Own Bolt",
		Types: []types.Card{types.Instant},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.CastForFree{
		Player:            game.ControllerReference(),
		Card:              game.CardReference{Kind: game.CardReferenceTarget},
		Zone:              zone.Graveyard,
		ExileOnResolution: true,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Instant}, Controller: game.ControllerYou}),
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("no spell on the stack after the free cast")
	}
	if !top.ExileOnResolution {
		t.Fatal("cast spell did not carry ExileOnResolution from the rider")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(targetID) {
		t.Fatal("resolved spell went to the graveyard despite the exile-instead rider")
	}
	if !g.Players[game.Player1].Exile.Contains(targetID) {
		t.Fatal("resolved spell was not exiled by the exile-instead rider")
	}
}
