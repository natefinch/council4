package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmpyrialArchangel is the card definition for Empyrial Archangel.
//
// Type: Creature — Angel
// Cost: {4}{G}{W}{W}{U}
//
// Oracle text:
//
//	Flying
//	Shroud (This creature can't be the target of spells or abilities.)
//	All damage that would be dealt to you is dealt to this creature instead.
var EmpyrialArchangel = newEmpyrialArchangel

func newEmpyrialArchangel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Empyrial Archangel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ShroudStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectRedirectDamageToSource,
							AffectedPlayer: game.PlayerYou,
						},
					},
				},
			},
			OracleText: `
			Flying
			Shroud (This creature can't be the target of spells or abilities.)
			All damage that would be dealt to you is dealt to this creature instead.
		`,
		},
	}
}
