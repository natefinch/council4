package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func playerGraveyardExileSource(g *game.Game, controller, targetPlayer game.PlayerID) {
	addEffectSpellToStack(g, controller, game.MoveCard{
		Player:      game.TargetPlayerReference(0),
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	}, []game.Target{{Kind: game.TargetPlayer, PlayerID: targetPlayer}})
}

// TestMoveCardExilesEntireTargetPlayerGraveyard verifies the player-zone group
// form moves every card in the chosen player's graveyard to that player's exile,
// preserving ownership, while leaving other players' graveyards untouched.
func TestMoveCardExilesEntireTargetPlayerGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	var targetCards []id.ID
	for _, name := range []string{"Gy One", "Gy Two", "Gy Three"} {
		targetCards = append(targetCards, addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{types.Creature},
		}}))
	}
	bystander := addCardToGraveyard(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bystander",
		Types: []types.Card{types.Instant},
	}})

	playerGraveyardExileSource(g, game.Player1, game.Player2)
	engine.resolveTopOfStack(g, &TurnLog{})

	if size := g.Players[game.Player2].Graveyard.Size(); size != 0 {
		t.Fatalf("target graveyard size = %d, want 0", size)
	}
	for _, cardID := range targetCards {
		if !g.Players[game.Player2].Exile.Contains(cardID) {
			t.Fatalf("card %v did not move to its owner's exile", cardID)
		}
	}
	if !g.Players[game.Player3].Graveyard.Contains(bystander) {
		t.Fatal("unrelated graveyard was disturbed")
	}
	if g.Players[game.Player3].Exile.Contains(bystander) {
		t.Fatal("unrelated card was exiled")
	}
}

// TestMoveCardPlayerGraveyardEmitsSingleZoneChangeBatch verifies every card
// exiled from the graveyard emits a zone-change event and that all of them share
// one non-zero SimultaneousID so the moves register as a single batch.
func TestMoveCardPlayerGraveyardEmitsSingleZoneChangeBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	var targetCards []id.ID
	for _, name := range []string{"Gy One", "Gy Two", "Gy Three"} {
		targetCards = append(targetCards, addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: name}}))
	}
	moved := map[id.ID]game.Event{}
	playerGraveyardExileSource(g, game.Player1, game.Player2)
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Graveyard && event.ToZone == zone.Exile {
			moved[event.CardID] = event
		}
	}
	if len(moved) != len(targetCards) {
		t.Fatalf("zone-change events = %d, want %d", len(moved), len(targetCards))
	}
	var batch id.ID
	for _, cardID := range targetCards {
		event, ok := moved[cardID]
		if !ok {
			t.Fatalf("missing zone-change event for %v", cardID)
		}
		if event.SimultaneousID == 0 {
			t.Fatalf("card %v zone change has no SimultaneousID", cardID)
		}
		if batch == 0 {
			batch = event.SimultaneousID
		} else if event.SimultaneousID != batch {
			t.Fatalf("card %v SimultaneousID = %v, want shared %v", cardID, event.SimultaneousID, batch)
		}
	}
}

// TestMoveCardPlayerGraveyardEmptyIsNoOp verifies that exiling an empty
// graveyard resolves cleanly and emits no zone-change events.
func TestMoveCardPlayerGraveyardEmptyIsNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	playerGraveyardExileSource(g, game.Player1, game.Player2)
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.ToZone == zone.Exile {
			t.Fatalf("unexpected zone-change event for empty graveyard: %#v", event)
		}
	}
	if size := g.Players[game.Player2].Exile.Size(); size != 0 {
		t.Fatalf("exile size = %d, want 0", size)
	}
}

// TestMoveCardPlayerGraveyardTargetsChosenPlayerOnly verifies the controller's
// own graveyard is untouched when an opponent is targeted, even though the
// controller also has cards in their graveyard.
func TestMoveCardPlayerGraveyardTargetsChosenPlayerOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	own := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Own Card"}})
	victim := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Victim"}})

	playerGraveyardExileSource(g, game.Player1, game.Player2)
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(own) {
		t.Fatal("controller's own graveyard card was exiled")
	}
	if !g.Players[game.Player2].Exile.Contains(victim) {
		t.Fatal("targeted player's card was not exiled")
	}
}
