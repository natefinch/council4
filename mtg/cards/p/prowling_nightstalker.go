package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ProwlingNightstalker is the card definition for Prowling Nightstalker.
//
// Type: Creature — Nightstalker
// Cost: {3}{B}
//
// Oracle text:
//
//	This creature can't be blocked except by black creatures.
var ProwlingNightstalker = newProwlingNightstalker()

func newProwlingNightstalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Prowling Nightstalker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nightstalker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedExceptBy,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionColor,
								Color: color.Black,
							},
						},
					},
				},
			},
			OracleText: `
			This creature can't be blocked except by black creatures.
		`,
		},
	}
}
