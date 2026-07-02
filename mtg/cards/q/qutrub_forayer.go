package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// QutrubForayer is the card definition for Qutrub Forayer.
//
// Type: Creature — Zombie Horror
// Cost: {2}{B}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Destroy target creature that was dealt damage this turn.
//	• Exile up to two target cards from a single graveyard.
var QutrubForayer = newQutrubForayer()

func newQutrubForayer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Qutrub Forayer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Horror},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Destroy target creature that was dealt damage this turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature that was dealt damage this turn",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, DealtDamageThisTurn: true}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Exile up to two target cards from a single graveyard.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets:    0,
										MaxTargets:    2,
										Constraint:    "up to two target cards from a single graveyard",
										Allow:         game.TargetAllowCard,
										TargetZone:    zone.Graveyard,
										Selection:     opt.Val(game.Selection{}),
										SameGraveyard: true,
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.MoveCard{
											Card:        game.CardReference{Kind: game.CardReferenceTarget},
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
										},
									},
									{
										Primitive: game.MoveCard{
											Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
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
			When this creature enters, choose one —
			• Destroy target creature that was dealt damage this turn.
			• Exile up to two target cards from a single graveyard.
		`,
		},
	}
}
