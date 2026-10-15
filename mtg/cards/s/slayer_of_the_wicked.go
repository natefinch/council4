package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlayerOfTheWicked is the card definition for Slayer of the Wicked.
//
// Type: Creature — Human Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	When this creature enters, you may destroy target Vampire, Werewolf, or Zombie.
var SlayerOfTheWicked = newSlayerOfTheWicked

func newSlayerOfTheWicked() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Slayer of the Wicked",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you may destroy target Vampire, Werewolf, or Zombie.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target Vampire, Werewolf, or Zombie",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Vampire, types.Werewolf, types.Zombie}}),
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
			When this creature enters, you may destroy target Vampire, Werewolf, or Zombie.
		`,
		},
	}
}
