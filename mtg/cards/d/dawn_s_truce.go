package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DawnSTruce is the card definition for Dawn's Truce.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Gift a card (You may promise an opponent a gift as you cast this spell. If you do, they draw a card before its other effects.)
//	You and permanents you control gain hexproof until end of turn. If the gift was promised, permanents you control also gain indestructible until end of turn.
var DawnSTruce = newDawnSTruce

func newDawnSTruce() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Dawn's Truce",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
									Primitive: game.Draw{
										Amount: game.Fixed(1),
										Player: game.GiftRecipientReference(),
									},
								},
							},
						}.Ability()},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectPlayerHexproof,
									AffectedPlayer: game.PlayerYou,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Hexproof,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Indestructible,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								GiftPromised: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Gift a card (You may promise an opponent a gift as you cast this spell. If you do, they draw a card before its other effects.)
			You and permanents you control gain hexproof until end of turn. If the gift was promised, permanents you control also gain indestructible until end of turn.
		`,
		},
	}
}
