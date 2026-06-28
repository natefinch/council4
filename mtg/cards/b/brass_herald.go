package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrassHerald is the card definition for Brass Herald.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	As this creature enters, choose a creature type.
//	When this creature enters, reveal the top four cards of your library. Put all creature cards of the chosen type revealed this way into your hand and the rest on the bottom of your library in any order.
//	Creatures of the chosen type get +1/+1.
var BrassHerald = newBrassHerald()

func newBrassHerald() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Brass Herald",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RevealTopPartition{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(4),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
									Remainder: game.DigRemainderLibraryBottom,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this creature enters, choose a creature type."),
			},
			OracleText: `
			As this creature enters, choose a creature type.
			When this creature enters, reveal the top four cards of your library. Put all creature cards of the chosen type revealed this way into your hand and the rest on the bottom of your library in any order.
			Creatures of the chosen type get +1/+1.
		`,
		},
	}
}
