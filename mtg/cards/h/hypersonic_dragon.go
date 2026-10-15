package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HypersonicDragon is the card definition for Hypersonic Dragon.
//
// Type: Creature — Dragon
// Cost: {3}{U}{R}
//
// Oracle text:
//
//	Flying, haste
//	You may cast sorcery spells as though they had flash. (You may cast them any time you could cast an instant.)
var HypersonicDragon = newHypersonicDragon

func newHypersonicDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Hypersonic Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
				game.StaticAbility{
					Text: "You may cast sorcery spells as though they had flash. (You may cast them any time you could cast an instant.)",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Sorcery},
						},
					},
				},
			},
			OracleText: `
			Flying, haste
			You may cast sorcery spells as though they had flash. (You may cast them any time you could cast an instant.)
		`,
		},
	}
}
