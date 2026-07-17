package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PadeemConsulOfInnovation is the card definition for Padeem, Consul of Innovation.
//
// Type: Legendary Creature — Vedalken Artificer
// Cost: {3}{U}
//
// Oracle text:
//
//	Artifacts you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)
//	At the beginning of your upkeep, if you control the artifact with the greatest mana value or tied for the greatest mana value, draw a card.
var PadeemConsulOfInnovation = newPadeemConsulOfInnovation

func newPadeemConsulOfInnovation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Padeem, Consul of Innovation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Vedalken, types.Artificer},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
							AddKeywords: []game.Keyword{
								game.Hexproof,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if you control the artifact with the greatest mana value or tied for the greatest mana value",
						InterveningCondition: opt.Val(game.Condition{
							ControlsGreatestManaValueInGroup: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Artifacts you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)
			At the beginning of your upkeep, if you control the artifact with the greatest mana value or tied for the greatest mana value, draw a card.
		`,
		},
	}
}
