package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// discardSubtypeUnionPermanent models Mary Read and Anne Bonny's payoff trigger:
// "Whenever you discard an Island, Pirate, or Vehicle card, ...". The card
// filter is a union of a land subtype (Island), a creature subtype (Pirate),
// and an artifact subtype (Vehicle), matched against the discarded card.
func discardSubtypeUnionPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addTriggeredPermanent(g, controller,
		&game.TriggerPattern{
			Event:  game.EventCardDiscarded,
			Player: game.TriggerPlayerYou,
			CardSelection: game.Selection{
				SubtypesAny: []types.Sub{types.Island, types.Pirate, types.Vehicle},
			},
		},
		[]game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}, nil)
}

func TestDiscardSubtypeUnionTriggerMatchesEachMember(t *testing.T) {
	cases := []struct {
		name    string
		def     *game.CardDef
		matched bool
	}{
		{
			name:    "land subtype",
			def:     &game.CardDef{CardFace: game.CardFace{Name: "Island", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Island}}},
			matched: true,
		},
		{
			name:    "creature subtype",
			def:     &game.CardDef{CardFace: game.CardFace{Name: "Pirate Crew", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human, types.Pirate}}},
			matched: true,
		},
		{
			name:    "artifact subtype",
			def:     &game.CardDef{CardFace: game.CardFace{Name: "Smuggler's Copter", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Vehicle}}},
			matched: true,
		},
		{
			name:    "non-member card",
			def:     &game.CardDef{CardFace: game.CardFace{Name: "Grizzly Bears", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Bear}}},
			matched: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			discardSubtypeUnionPermanent(g, game.Player1)
			cardID := addCardToHand(g, game.Player1, tc.def)

			if !discardCardFromHand(g, game.Player1, cardID) {
				t.Fatal("discardCardFromHand() = false, want true")
			}
			placed := engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if placed != tc.matched {
				t.Fatalf("trigger placed = %v, want %v", placed, tc.matched)
			}
		})
	}
}
