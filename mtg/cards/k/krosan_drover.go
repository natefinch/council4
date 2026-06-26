package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KrosanDrover is the card definition for Krosan Drover.
//
// Type: Creature — Elf
// Cost: {3}{G}
//
// Oracle text:
//
//	Creature spells you cast with mana value 6 or greater cost {2} less to cast.
var KrosanDrover = newKrosanDrover()

func newKrosanDrover() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Krosan Drover",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Creature}, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 6})},
								GenericReduction: 2,
							},
						},
					},
				},
			},
			OracleText: `
			Creature spells you cast with mana value 6 or greater cost {2} less to cast.
		`,
		},
	}
}
