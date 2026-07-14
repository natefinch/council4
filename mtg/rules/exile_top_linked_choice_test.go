package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestExileTopLinkedAnyNumberChoiceUsesOnlyExiledBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	oldExile := addCardToExile(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Old Creature", Types: []types.Card{types.Creature}},
	})
	creature := addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Creature", Types: []types.Card{types.Creature}},
	})
	land := addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}},
	})
	instant := addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Instant", Types: []types.Card{types.Instant}},
	})
	const link game.LinkedKey = "exiled-top-cards"
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.ExileTopOfLibrary{
			Amount:        game.Fixed(3),
			Player:        game.ControllerReference(),
			PublishLinked: link,
		}},
		{Primitive: game.ChooseFromZone{
			Player:      game.ControllerReference(),
			SourceZone:  zone.Exile,
			Filter:      game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
			Count:       game.ChooseAnyNumber,
			Destination: game.ChooseDestination{Zone: zone.Battlefield},
			Riders:      game.ChooseRiders{FromLinked: link},
		}},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: scriptedChoiceAgent{answer: []int{0, 1}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, cardID := range []id.ID{creature, land} {
		if permanentForCard(g, cardID) == nil {
			t.Fatalf("linked card %v did not enter the battlefield", cardID)
		}
	}
	if !g.Players[game.Player1].Exile.Contains(instant) {
		t.Fatal("linked nonmatching instant left exile")
	}
	if !g.Players[game.Player1].Exile.Contains(oldExile) {
		t.Fatal("unlinked preexisting exiled card was eligible")
	}
}
