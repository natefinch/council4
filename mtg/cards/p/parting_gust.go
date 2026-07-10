package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PartingGust is the card definition for Parting Gust.
//
// Type: Instant
// Cost: {W}{W}
//
// Oracle text:
//
//	Gift a tapped Fish (You may promise an opponent a gift as you cast this spell. If you do, they create a tapped 1/1 blue Fish creature token before its other effects.)
//	Exile target nontoken creature. If the gift wasn't promised, return that card to the battlefield under its owner's control with a +1/+1 counter on it at the beginning of the next end step.
var PartingGust = newPartingGust

func newPartingGust() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Parting Gust",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.GiftKeyword{Delivery: game.Mode{
							Sequence: []game.Instruction{
								{
									Primitive: game.CreateToken{
										Amount:      game.Fixed(1),
										Source:      game.TokenDef(partingGustToken),
										Recipient:   opt.Val(game.GiftRecipientReference()),
										EntryTapped: true,
									},
								},
							},
						}.Ability()},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nontoken creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, NonToken: true}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Exile{
							Object:         game.TargetPermanentReference(0),
							ExileLinkedKey: game.LinkedKey("delayed-blink-1"),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.PutOnBattlefield{
												Source:        game.LinkedBattlefieldSource(game.LinkedKey("delayed-blink-1")),
												EntryCounters: []game.CounterPlacement{game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}},
											},
										},
									},
								}.Ability(),
							},
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								GiftPromised: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Gift a tapped Fish (You may promise an opponent a gift as you cast this spell. If you do, they create a tapped 1/1 blue Fish creature token before its other effects.)
			Exile target nontoken creature. If the gift wasn't promised, return that card to the battlefield under its owner's control with a +1/+1 counter on it at the beginning of the next end step.
		`,
		},
	}
}

var partingGustToken = newPartingGustToken()

func newPartingGustToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Fish",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fish},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
