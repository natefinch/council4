package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AgentOfTheIronThrone is the card definition for Agent of the Iron Throne.
//
// Type: Legendary Enchantment — Background
// Cost: {2}{B}
//
// Oracle text:
//
//	Commander creatures you own have "Whenever an artifact or creature you control is put into a graveyard from the battlefield, each opponent loses 1 life."
var AgentOfTheIronThrone = newAgentOfTheIronThrone

func newAgentOfTheIronThrone() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Agent of the Iron Throne",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCommander: true}),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:            game.EventZoneChanged,
											Controller:       game.TriggerControllerYou,
											MatchFromZone:    true,
											FromZone:         zone.Battlefield,
											MatchToZone:      true,
											ToZone:           zone.Graveyard,
											SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}},
										},
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.LoseLife{
													Amount:      game.Fixed(1),
													PlayerGroup: game.OpponentsReference(),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Commander creatures you own have "Whenever an artifact or creature you control is put into a graveyard from the battlefield, each opponent loses 1 life."
		`,
		},
	}
}
