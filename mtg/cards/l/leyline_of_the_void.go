package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeylineOfTheVoid is the card definition for Leyline of the Void.
//
// Type: Enchantment
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	If this card is in your opening hand, you may begin the game with it on the battlefield.
//	If a card would be put into an opponent's graveyard from anywhere, exile it instead.
var LeylineOfTheVoid = newLeylineOfTheVoid

func newLeylineOfTheVoid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Leyline of the Void",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					BeginsGameOnBattlefield: true,
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectReplacement("If a card would be put into an opponent's graveyard from anywhere, exile it instead.", game.TriggerControllerOpponent, game.TriggerControllerAny, false),
			},
			OracleText: `
			If this card is in your opening hand, you may begin the game with it on the battlefield.
			If a card would be put into an opponent's graveyard from anywhere, exile it instead.
		`,
		},
	}
}
