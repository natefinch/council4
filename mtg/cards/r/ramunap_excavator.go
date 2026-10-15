package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RamunapExcavator is the card definition for Ramunap Excavator.
//
// Type: Creature — Snake Cleric
// Cost: {2}{G}
//
// Oracle text:
//
//	You may play lands from your graveyard.
var RamunapExcavator = newRamunapExcavator

func newRamunapExcavator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ramunap Excavator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.PlayLandsFromGraveyardStaticBody,
			},
			OracleText: `
			You may play lands from your graveyard.
		`,
		},
	}
}
