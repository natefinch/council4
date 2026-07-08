package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TreasureHunt is the card definition for Treasure Hunt.
//
// Type: Sorcery
// Cost: {1}{U}
//
// Oracle text:
//
//	Reveal cards from the top of your library until you reveal a nonland card, then put all cards revealed this way into your hand.
var TreasureHunt = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Treasure Hunt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RevealUntil{
							Player:      game.ControllerReference(),
							Until:       game.Selection{ExcludedTypes: []types.Card{types.Land}},
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal cards from the top of your library until you reveal a nonland card, then put all cards revealed this way into your hand.
		`,
		},
	}
}
