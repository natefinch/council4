package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrchardWarden is the card definition for Orchard Warden.
//
// Type: Creature — Treefolk Shaman
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Whenever another Treefolk creature you control enters, you may gain life equal to that creature's toughness.
var OrchardWarden = newOrchardWarden()

func newOrchardWarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Orchard Warden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Shaman},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Treefolk")}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectToughness,
										Multiplier: 1,
										Object:     game.EventPermanentReference(),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever another Treefolk creature you control enters, you may gain life equal to that creature's toughness.
		`,
		},
	}
}
