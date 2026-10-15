package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MithrilCoat is the card definition for Mithril Coat.
//
// Type: Legendary Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Flash
//	Indestructible
//	When Mithril Coat enters, attach it to target legendary creature you control.
//	Equipped creature has indestructible.
//	Equip {3}
var MithrilCoat = newMithrilCoat

func newMithrilCoat() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mithril Coat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Text: "Equipped creature has indestructible.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Indestructible},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Mithril Coat enters, attach it to target legendary creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
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
										Constraint: "target legendary creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}, Controller: game.ControllerYou}),
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
			Flash
			Indestructible
			When Mithril Coat enters, attach it to target legendary creature you control.
			Equipped creature has indestructible.
			Equip {3}
		`,
		},
	}
}
