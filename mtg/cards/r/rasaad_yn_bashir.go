package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RasaadYnBashir is the card definition for Rasaad yn Bashir.
//
// Type: Legendary Creature — Human Monk
// Cost: {2}{W}
//
// Oracle text:
//
//	Each creature you control assigns combat damage equal to its toughness rather than its power.
//	Whenever Rasaad yn Bashir attacks, if you have the initiative, double the toughness of each creature you control until end of turn.
//	Choose a Background (You can have a Background as a second commander.)
var RasaadYnBashir = newRasaadYnBashir

func newRasaadYnBashir() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rasaad yn Bashir",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Monk},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectAssignCombatDamageUsingToughness,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
						},
					},
				},
				game.ChooseABackgroundStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you have the initiative",
						InterveningCondition: opt.Val(game.Condition{
							ControllerHasInitiative: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:           game.LayerPowerToughnessModify,
											Group:           game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											DoubleToughness: true,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Each creature you control assigns combat damage equal to its toughness rather than its power.
			Whenever Rasaad yn Bashir attacks, if you have the initiative, double the toughness of each creature you control until end of turn.
			Choose a Background (You can have a Background as a second commander.)
		`,
		},
	}
}
