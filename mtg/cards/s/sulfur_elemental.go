package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SulfurElemental is the card definition for Sulfur Elemental.
//
// Type: Creature — Elemental
// Cost: {2}{R}
//
// Oracle text:
//
//	Flash
//	Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
//	White creatures get +1/-1.
var SulfurElemental = newSulfurElemental

func newSulfurElemental() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Sulfur Elemental",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.SplitSecondStaticBody,
				game.StaticAbility{
					Text: "White creatures get +1/-1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.White}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: -1,
						},
					},
				},
			},
			OracleText: `
			Flash
			Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
			White creatures get +1/-1.
		`,
		},
	}
}
