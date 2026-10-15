package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoltenMonstrosity is the card definition for Molten Monstrosity.
//
// Type: Creature — Hellion
// Cost: {7}{R}
//
// Oracle text:
//
//	This spell costs {X} less to cast, where X is the greatest power among creatures you control.
//	Trample
var MoltenMonstrosity = newMoltenMonstrosity

func newMoltenMonstrosity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Molten Monstrosity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {X} less to cast, where X is the greatest power among creatures you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								DynamicReduction: &game.DynamicAmount{
									Kind:        game.DynamicAmountGreatestPowerInGroup,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
								},
							},
						},
					},
				},
				game.TrampleStaticBody,
			},
			OracleText: `
			This spell costs {X} less to cast, where X is the greatest power among creatures you control.
			Trample
		`,
		},
	}
}
