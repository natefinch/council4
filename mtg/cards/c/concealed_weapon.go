package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConcealedWeapon is the card definition for Concealed Weapon.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +3/+0.
//	Disguise {2}{R} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
//	When this Equipment is turned face up, attach it to target creature you control.
//	Equip {1}{R}
var ConcealedWeapon = newConcealedWeapon

func newConcealedWeapon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Concealed Weapon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +3/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 3,
						},
					},
				},
				game.StaticAbility{
					Text: "Disguise {2}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.DisguiseKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.R,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1), cost.R}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this Equipment is turned face up, attach it to target creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentTurnedFaceUp,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Attach{
											Attachment: game.EventPermanentReference(),
											Target:     game.TargetPermanentReference(0),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Equipped creature gets +3/+0.
			Disguise {2}{R} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
			When this Equipment is turned face up, attach it to target creature you control.
			Equip {1}{R}
		`,
		},
	}
}
