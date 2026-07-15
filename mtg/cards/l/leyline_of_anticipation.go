package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeylineOfAnticipation is the card definition for Leyline of Anticipation.
//
// Type: Enchantment
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	If this card is in your opening hand, you may begin the game with it on the battlefield.
//	You may cast spells as though they had flash.
var LeylineOfAnticipation = newLeylineOfAnticipation

func newLeylineOfAnticipation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Leyline of Anticipation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					BeginsGameOnBattlefield: true,
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							AffectedPlayer: game.PlayerYou,
						},
					},
				},
			},
			OracleText: `
			If this card is in your opening hand, you may begin the game with it on the battlefield.
			You may cast spells as though they had flash.
		`,
		},
	}
}
