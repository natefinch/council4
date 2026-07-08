package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JirinaDauntlessGeneral is the card definition for Jirina, Dauntless General.
//
// Type: Legendary Creature — Human Soldier
// Cost: {W}{B}
//
// Oracle text:
//
//	When Jirina enters, exile target player's graveyard.
//	Sacrifice Jirina: Humans you control gain hexproof and indestructible until end of turn.
var JirinaDauntlessGeneral = newJirinaDauntlessGeneral

func newJirinaDauntlessGeneral() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Jirina, Dauntless General",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice Jirina: Humans you control gain hexproof and indestructible until end of turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice Jirina",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Hexproof,
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player's graveyard",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Player:      game.TargetPlayerReference(0),
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Jirina enters, exile target player's graveyard.
			Sacrifice Jirina: Humans you control gain hexproof and indestructible until end of turn.
		`,
		},
	}
}
