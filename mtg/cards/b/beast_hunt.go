package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BeastHunt is the card definition for Beast Hunt.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Reveal the top three cards of your library. Put all creature cards revealed this way into your hand and the rest into your graveyard.
var BeastHunt = newBeastHunt

func newBeastHunt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Beast Hunt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RevealTopPartition{
							Player:    game.ControllerReference(),
							Amount:    game.Fixed(3),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top three cards of your library. Put all creature cards revealed this way into your hand and the rest into your graveyard.
		`,
		},
	}
}
