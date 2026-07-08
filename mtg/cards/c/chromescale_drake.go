package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChromescaleDrake is the card definition for Chromescale Drake.
//
// Type: Creature — Drake
// Cost: {6}{U}{U}{U}
//
// Oracle text:
//
//	Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
//	Flying
//	When this creature enters, reveal the top three cards of your library. Put all artifact cards revealed this way into your hand and the rest into your graveyard.
var ChromescaleDrake = newChromescaleDrake

func newChromescaleDrake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Chromescale Drake",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
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
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RevealTopPartition{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(3),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
			Flying
			When this creature enters, reveal the top three cards of your library. Put all artifact cards revealed this way into your hand and the rest into your graveyard.
		`,
		},
	}
}
