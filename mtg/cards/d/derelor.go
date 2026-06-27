package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Derelor is the card definition for Derelor.
//
// Type: Creature — Thrull
// Cost: {3}{B}
//
// Oracle text:
//
//	Black spells you cast cost {B} more to cast.
var Derelor = newDerelor()

func newDerelor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Derelor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Thrull},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{ColorsAny: []color.Color{color.Black}},
								ColoredIncrease: []mana.Color{mana.B},
							},
						},
					},
				},
			},
			OracleText: `
			Black spells you cast cost {B} more to cast.
		`,
		},
	}
}
