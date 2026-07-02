package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DailyBugleReporters is the card definition for Daily Bugle Reporters.
//
// Type: Creature — Human Citizen
// Cost: {3}{W}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Puff Piece — Put a +1/+1 counter on each of up to two target creatures.
//	• Investigative Journalism — Return target creature card with mana value 2 or less from your graveyard to your hand.
var DailyBugleReporters = newDailyBugleReporters()

func newDailyBugleReporters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Daily Bugle Reporters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Citizen},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
								Text: "Puff Piece — Put a +1/+1 counter on each of up to two target creatures.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 0,
										MaxTargets: 2,
										Constraint: "up to two target creatures",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(1),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "Investigative Journalism — Return target creature card with mana value 2 or less from your graveyard to your hand.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature card with mana value 2 or less from your graveyard",
										Allow:      game.TargetAllowCard,
										TargetZone: zone.Graveyard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.MoveCard{
											Card:        game.CardReference{Kind: game.CardReferenceTarget},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			• Puff Piece — Put a +1/+1 counter on each of up to two target creatures.
			• Investigative Journalism — Return target creature card with mana value 2 or less from your graveyard to your hand.
		`,
		},
	}
}
