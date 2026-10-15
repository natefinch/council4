package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArlinnSWolf is the card definition for Arlinn's Wolf.
//
// Type: Creature — Wolf
// Cost: {2}{G}
//
// Oracle text:
//
//	This creature can't be blocked by creatures with power 2 or less.
var ArlinnSWolf = newArlinnSWolf

func newArlinnSWolf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Arlinn's Wolf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't be blocked by creatures with power 2 or less.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			OracleText: `
			This creature can't be blocked by creatures with power 2 or less.
		`,
		},
	}
}
