package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IdyllicGrange is the card definition for Idyllic Grange.
//
// Type: Land — Plains
//
// Oracle text:
//
//	({T}: Add {W}.)
//	This land enters tapped unless you control three or more other Plains.
//	When this land enters untapped, put a +1/+1 counter on target creature you control.
var IdyllicGrange = newIdyllicGrange

func newIdyllicGrange() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Idyllic Grange",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters untapped, put a +1/+1 counter on target creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{Tapped: game.TriFalse},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control three or more other Plains.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Plains")}, ExcludeSource: true},
						MinCount:  3,
					}),
				}),
			},
			OracleText: `
			({T}: Add {W}.)
			This land enters tapped unless you control three or more other Plains.
			When this land enters untapped, put a +1/+1 counter on target creature you control.
		`,
		},
	}
}
