package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DaybreakChimera is the card definition for Daybreak Chimera.
//
// Type: Creature — Chimera
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	This spell costs {X} less to cast, where X is your devotion to white. (Each {W} in the mana costs of permanents you control counts toward your devotion to white.)
//	Flying
var DaybreakChimera = newDaybreakChimera

func newDaybreakChimera() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Daybreak Chimera",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Chimera},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {X} less to cast, where X is your devotion to white. (Each {W} in the mana costs of permanents you control counts toward your devotion to white.)",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								DynamicReduction: &game.DynamicAmount{
									Kind:        game.DynamicAmountDevotion,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Colors:      []color.Color{color.White},
								},
							},
						},
					},
				},
				game.FlyingStaticBody,
			},
			OracleText: `
			This spell costs {X} less to cast, where X is your devotion to white. (Each {W} in the mana costs of permanents you control counts toward your devotion to white.)
			Flying
		`,
		},
	}
}
