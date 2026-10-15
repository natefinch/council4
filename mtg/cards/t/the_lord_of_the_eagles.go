package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheLordOfTheEagles is the card definition for The Lord of the Eagles.
//
// Type: Legendary Creature — Bird Noble
// Cost: {7}{U}{U}
//
// Oracle text:
//
//	Flash
//	This spell costs {X} less to cast, where X is the total power of creatures you control with flying.
//	Flying
var TheLordOfTheEagles = newTheLordOfTheEagles

func newTheLordOfTheEagles() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "The Lord of the Eagles",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Bird, types.Noble},
			Power:      opt.Val(game.PT{Value: 8}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "This spell costs {X} less to cast, where X is the total power of creatures you control with flying.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								DynamicReduction: &game.DynamicAmount{
									Kind:        game.DynamicAmountTotalPowerInGroup,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Keyword: game.Flying}),
								},
							},
						},
					},
				},
				game.FlyingStaticBody,
			},
			OracleText: `
			Flash
			This spell costs {X} less to cast, where X is the total power of creatures you control with flying.
			Flying
		`,
		},
	}
}
