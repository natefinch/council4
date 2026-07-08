package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoulbladeCorrupter is the card definition for Soulblade Corrupter.
//
// Type: Creature — Human Warrior
// Cost: {4}{B}
//
// Oracle text:
//
//	Partner with Soulblade Renewer (When this creature enters, target player may put Soulblade Renewer into their hand from their library, then shuffle.)
//	Deathtouch
//	Whenever a creature with a +1/+1 counter on it attacks one of your opponents, that creature gains deathtouch until end of turn.
var SoulbladeCorrupter = newSoulbladeCorrupter

func newSoulbladeCorrupter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Soulblade Corrupter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
				game.DeathtouchStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Player:           game.TriggerPlayerOpponent,
							AttackRecipient:  game.AttackRecipientPlayer,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.EventPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Deathtouch,
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
			OracleText: `
			Partner with Soulblade Renewer (When this creature enters, target player may put Soulblade Renewer into their hand from their library, then shuffle.)
			Deathtouch
			Whenever a creature with a +1/+1 counter on it attacks one of your opponents, that creature gains deathtouch until end of turn.
		`,
		},
	}
}
