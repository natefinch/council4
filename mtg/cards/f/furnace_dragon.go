package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FurnaceDragon is the card definition for Furnace Dragon.
//
// Type: Creature — Dragon
// Cost: {6}{R}{R}{R}
//
// Oracle text:
//
//	Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
//	Flying
//	When this creature enters, if you cast it from your hand, exile all artifacts.
var FurnaceDragon = newFurnaceDragon

func newFurnaceDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Furnace Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou},
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
						InterveningIf: "if you cast it from your hand",
						InterveningIfEventPermanentWasCastFromControllerHand: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
			Flying
			When this creature enters, if you cast it from your hand, exile all artifacts.
		`,
		},
	}
}
