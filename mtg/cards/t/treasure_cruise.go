package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TreasureCruise is the card definition for Treasure Cruise.
//
// Type: Sorcery
// Cost: {7}{U}
//
// Oracle text:
//
//	Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
//	Draw three cards.
var TreasureCruise = newTreasureCruise

func newTreasureCruise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Treasure Cruise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.DelveStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
			Draw three cards.
		`,
		},
	}
}
