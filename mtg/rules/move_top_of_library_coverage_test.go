package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestMoveTopOfLibraryCoversUnwrittenDestinationRiderCombinations is the payoff
// test for collapsing Mill and ExileTopOfLibrary into one
// destination-parameterized primitive (game.MoveTopOfLibrary).
//
// The two old primitives overlapped heavily: both carried Amount, a player or
// player-group referent, and PublishLinked. ExileTopOfLibrary additionally owned
// Counter and FaceDown, but those are exile-only runtime state (Game.ExileCounters
// holds only cards in exile, and only the exile zone records face-down status),
// so validation still gates them to the exile destination — collapsing the types
// does not make them honorable at the graveyard.
//
// The one rider combination the collapse genuinely makes newly reachable is a
// group-form PublishLinked at the graveyard destination. The split Mill primitive
// rejected PublishLinked for its group form (its handler ignored it and its
// validator refused it), while ExileTopOfLibrary published across the group. The
// single unified handler now publishes for the group form at either destination,
// so "each player mills N cards, remembering exactly those cards" resolves with no
// new runtime code.
//
// The graveyard-group-publish subtest fails if the handler is re-specialized to
// publish only at the exile destination (or only for the single-player form); the
// exile-group-publish subtest documents the parity the collapse guarantees.
func TestMoveTopOfLibraryCoversUnwrittenDestinationRiderCombinations(t *testing.T) {
	const link game.LinkedKey = "moved-cards"

	t.Run("graveyard group mill publishes each milled card", func(t *testing.T) {
		// "Each player mills two cards" while remembering exactly those cards.
		// Group-form PublishLinked at the graveyard destination had no
		// expressible primitive before the collapse: Mill's group form could not
		// publish a linked key.
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, distributiveDestroySagaDef())
		obj := linkedSourceObject(source)

		milled := make(map[id.ID]bool)
		for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
			addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
			for _, name := range []string{"MidA", "Top"} {
				milled[addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: name}})] = true
			}
		}

		engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
			Primitive: game.MoveTopOfLibrary{
				Destination:   zone.Graveyard,
				Amount:        game.Fixed(2),
				PlayerGroup:   game.AllPlayersReference(),
				PublishLinked: link,
			},
		}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		linked := linkedObjects(g, linkedObjectSourceKey(g, obj, string(link)))
		if len(linked) != 8 {
			t.Fatalf("linked objects = %d, want the 8 milled cards (2 per player)", len(linked))
		}
		for _, ref := range linked {
			if !milled[ref.CardID] {
				t.Fatalf("linked card %v was not one of the milled cards", ref.CardID)
			}
			if !g.Players[cardOwnerForTest(g, ref.CardID)].Graveyard.Contains(ref.CardID) {
				t.Fatalf("published card %v did not reach a graveyard", ref.CardID)
			}
		}
	})

	t.Run("exile group publishes each exiled card", func(t *testing.T) {
		// Parity check: the exile destination still publishes for the group form,
		// now through the same unified handler as the mill above.
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, distributiveDestroySagaDef())
		obj := linkedSourceObject(source)

		exiled := make(map[id.ID]bool)
		for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
			exiled[addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})] = true
		}

		engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
			Primitive: game.MoveTopOfLibrary{
				Destination:   zone.Exile,
				Amount:        game.Fixed(1),
				PlayerGroup:   game.AllPlayersReference(),
				PublishLinked: link,
			},
		}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		linked := linkedObjects(g, linkedObjectSourceKey(g, obj, string(link)))
		if len(linked) != 4 {
			t.Fatalf("linked objects = %d, want the 4 exiled cards", len(linked))
		}
		for _, ref := range linked {
			if !exiled[ref.CardID] {
				t.Fatalf("linked card %v was not one of the exiled cards", ref.CardID)
			}
		}
	})
}

// cardOwnerForTest returns the owner of a card instance, used to locate the
// graveyard a published milled card should rest in.
func cardOwnerForTest(g *game.Game, cardID id.ID) game.PlayerID {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return game.Player1
	}
	return card.Owner
}
