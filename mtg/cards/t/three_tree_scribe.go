package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThreeTreeScribe is the card definition for Three Tree Scribe.
//
// Type: Creature — Frog Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	Whenever this creature or another creature you control leaves the battlefield without dying, put a +1/+1 counter on target creature you control.
var ThreeTreeScribe = newThreeTreeScribe

func newThreeTreeScribe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Three Tree Scribe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Frog, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature or another creature you control leaves the battlefield without dying, put a +1/+1 counter on target creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventZoneChanged,
							Controller:             game.TriggerControllerYou,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}},
							SubjectSelectionOrSelf: true,
							MatchFromZone:          true,
							FromZone:               zone.Battlefield,
							ToZone:                 zone.Graveyard,
							ExcludeToZone:          true,
							StackObjectKind:        game.StackSpell,
							CounterKind:            counter.PlusOnePlusOne,
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
			OracleText: `
			Whenever this creature or another creature you control leaves the battlefield without dying, put a +1/+1 counter on target creature you control.
		`,
		},
	}
}
