package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RainSlickedCopse is the card definition for Rain-Slicked Copse.
//
// Type: Land — Forest Island
//
// Oracle text:
//
//	({T}: Add {G} or {U}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var RainSlickedCopse = newRainSlickedCopse

func newRainSlickedCopse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:     "Rain-Slicked Copse",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Island},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {G} or {U}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
