package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AvatarOfDiscord is the card definition for Avatar of Discord.
//
// Type: Creature — Avatar
// Cost: {B/R}{B/R}{B/R}
//
// Oracle text:
//
//	({B/R} can be paid with either {B} or {R}.)
//	Flying
//	When this creature enters, sacrifice it unless you discard two cards.
var AvatarOfDiscord = newAvatarOfDiscord

func newAvatarOfDiscord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Avatar of Discord",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.R),
				cost.HybridMana(mana.B, mana.R),
				cost.HybridMana(mana.B, mana.R),
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Avatar},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
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
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Discard two cards?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:   cost.AdditionalDiscard,
												Text:   "discard two cards",
												Amount: 2,
												Source: zone.Hand,
											},
										},
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			({B/R} can be paid with either {B} or {R}.)
			Flying
			When this creature enters, sacrifice it unless you discard two cards.
		`,
		},
	}
}
