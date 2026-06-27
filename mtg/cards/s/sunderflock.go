package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sunderflock is the card definition for Sunderflock.
//
// Type: Creature — Elemental
// Cost: {7}{U}{U}
//
// Oracle text:
//
//	This spell costs {X} less to cast, where X is the greatest mana value among Elementals you control.
//	Flying
//	When this creature enters, if you cast it, return all non-Elemental creatures to their owners' hands.
var Sunderflock = newSunderflock()

func newSunderflock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sunderflock",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								DynamicReduction: &game.DynamicAmount{
									Kind:       game.DynamicAmountGreatestManaValueInGroup,
									Multiplier: 1,
									Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elemental")}, Controller: game.ControllerYou}),
								},
							},
						},
					},
				},
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you cast it",
						InterveningIfEventPermanentWasCastByController: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Elemental")}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This spell costs {X} less to cast, where X is the greatest mana value among Elementals you control.
			Flying
			When this creature enters, if you cast it, return all non-Elemental creatures to their owners' hands.
		`,
		},
	}
}
