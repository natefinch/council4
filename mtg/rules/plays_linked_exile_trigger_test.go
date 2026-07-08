package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestPlaysLinkedExileCardTriggerMatching proves the PlaysLinkedExileCard
// pattern (Prowl's "whenever a player plays a card exiled with Prowl") matches
// an EventCardPlayedFromExile only when the played card belongs to the trigger
// source's linked-exile pool.
func TestPlaysLinkedExileCardTriggerMatching(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	const prowlLink = game.LinkedKey("prowl-exile")
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Prowl Source",
		Types: []types.Card{types.Creature},
	}})
	linkedCard := addCardToExile(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Linked"}})
	unlinkedCard := addCardToExile(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Unlinked"}})
	rememberLinkedObject(g, game.LinkedObjectKey{SourceID: source.CardInstanceID, LinkID: string(prowlLink)}, game.LinkedObjectRef{CardID: linkedCard})

	pattern := &game.TriggerPattern{
		Event:                game.EventCardPlayedFromExile,
		PlaysLinkedExileCard: prowlLink,
	}

	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:       game.EventCardPlayedFromExile,
		Controller: game.Player2,
		Player:     game.Player2,
		CardID:     linkedCard,
	}) {
		t.Fatal("trigger did not match a played card in the source's linked-exile pool")
	}
	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:       game.EventCardPlayedFromExile,
		Controller: game.Player2,
		Player:     game.Player2,
		CardID:     unlinkedCard,
	}) {
		t.Fatal("trigger matched a played card outside the source's linked-exile pool")
	}
}

// TestPlaysLinkedExileCardTriggerDrawsAndFires proves the plays-a-linked-exiled-
// card trigger goes on the stack and resolves its payoff when any player plays a
// card that was exiled with the source. It exercises the real detection and
// resolution path via the runtime event emission helper.
func TestPlaysLinkedExileCardTriggerDrawsAndFires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	const prowlLink = game.LinkedKey("prowl-exile")

	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventCardPlayedFromExile,
		PlaysLinkedExileCard: prowlLink,
	}, []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})

	playedCard := addCardToExile(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Played"}})
	rememberLinkedObject(g, game.LinkedObjectKey{SourceID: source.CardInstanceID, LinkID: string(prowlLink)}, game.LinkedObjectRef{CardID: playedCard})

	emitCardPlayedFromExileEvent(g, game.Player2, playedCard, zone.Exile)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("plays-a-linked-exiled-card trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("controller hand size = %d, want the trigger to draw one card", got)
	}
}
