package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HulkAlwaysAngry is the card definition for Hulk, Always Angry.
//
// Type: Legendary Creature — Gamma Berserker Hero
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	Trample
//	When Hulk enters, destroy all artifacts.
//	Hulk attacks each combat if able.
var HulkAlwaysAngry = newHulkAlwaysAngry

func newHulkAlwaysAngry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hulk, Always Angry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Gamma"), types.Berserker, types.Hero},
			Power:      opt.Val(game.PT{Value: 9}),
			Toughness:  opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.MustAttackStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Hulk enters, destroy all artifacts.",
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
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
			Trample
			When Hulk enters, destroy all artifacts.
			Hulk attacks each combat if able.
		`,
		},
	}
}
