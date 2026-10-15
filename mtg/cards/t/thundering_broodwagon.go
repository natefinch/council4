package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThunderingBroodwagon is the card definition for Thundering Broodwagon.
//
// Type: Artifact — Vehicle
// Cost: {2}{B}{B}{G}{G}
//
// Oracle text:
//
//	Reach, menace
//	When this Vehicle enters, destroy target nonland permanent an opponent controls with mana value 4 or less.
//	Crew 3
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var ThunderingBroodwagon = newThunderingBroodwagon

func newThunderingBroodwagon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Thundering Broodwagon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.MenaceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(3),
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this Vehicle enters, destroy target nonland permanent an opponent controls with mana value 4 or less.",
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
										Constraint: "target nonland permanent an opponent controls with mana value 4 or less",
										Allow:      game.TargetAllowPermanent,
										Selection: opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerOpponent, ManaValue: opt.Val(compare.Int{
											Op:    compare.LessOrEqual,
											Value: 4,
										})}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
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
			Reach, menace
			When this Vehicle enters, destroy target nonland permanent an opponent controls with mana value 4 or less.
			Crew 3
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
