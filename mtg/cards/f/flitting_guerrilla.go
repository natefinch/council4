package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FlittingGuerrilla is the card definition for Flitting Guerrilla.
//
// Type: Creature — Faerie Rogue
// Cost: {2}{B}
//
// Oracle text:
//
//	Flying
//	When this creature dies, each player mills two cards. Then you may exile this card. When you do, put target creature or battle card from your graveyard on top of your library. (To mill two cards, a player puts the top two cards of their library into their graveyard.)
var FlittingGuerrilla = newFlittingGuerrilla

func newFlittingGuerrilla() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Flitting Guerrilla",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount:      game.Fixed(2),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
							{
								Primitive: game.Exile{
									Object: game.SourceCardPermanentReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Targets: []game.TargetSpec{
												game.TargetSpec{
													MinTargets: 1,
													MaxTargets: 1,
													Constraint: "target creature or battle card from your graveyard",
													Allow:      game.TargetAllowCard,
													TargetZone: zone.Graveyard,
													Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Battle}, Controller: game.ControllerYou}),
												},
											},
											Sequence: []game.Instruction{
												{
													Primitive: game.MoveCard{
														Card:        game.CardReference{Kind: game.CardReferenceTarget},
														FromZone:    zone.Graveyard,
														Destination: zone.Library,
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this creature dies, each player mills two cards. Then you may exile this card. When you do, put target creature or battle card from your graveyard on top of your library. (To mill two cards, a player puts the top two cards of their library into their graveyard.)
		`,
		},
	}
}
