package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThaliaGuardianOfThraben is the card definition for Thalia, Guardian of Thraben.
//
// Type: Legendary Creature — Human Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	First strike
//	Noncreature spells cost {1} more to cast.
var ThaliaGuardianOfThraben = newThaliaGuardianOfThraben()

func newThaliaGuardianOfThraben() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Thalia, Guardian of Thraben",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
								GenericIncrease: 1,
							},
						},
					},
				},
			},
			OracleText: `
			First strike
			Noncreature spells cost {1} more to cast.
		`,
		},
	}
}
