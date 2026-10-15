package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NaruMehaMasterWizard is the card definition for Naru Meha, Master Wizard.
//
// Type: Legendary Creature — Human Wizard
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Flash
//	When Naru Meha enters, copy target instant or sorcery spell you control. You may choose new targets for the copy.
//	Other Wizards you control get +1/+1.
var NaruMehaMasterWizard = newNaruMehaMasterWizard

func newNaruMehaMasterWizard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Naru Meha, Master Wizard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "Other Wizards you control get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Wizard")}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Naru Meha enters, copy target instant or sorcery spell you control. You may choose new targets for the copy.",
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
										Constraint: "target instant or sorcery spell you control",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
											StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
											Controller:        game.ControllerYou,
										},
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CopyStackObject{
											Object:              game.TargetStackObjectReference(0),
											Count:               1,
											MayChooseNewTargets: true,
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
			When Naru Meha enters, copy target instant or sorcery spell you control. You may choose new targets for the copy.
			Other Wizards you control get +1/+1.
		`,
		},
	}
}
