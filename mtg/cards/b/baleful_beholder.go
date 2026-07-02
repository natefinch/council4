package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BalefulBeholder is the card definition for Baleful Beholder.
//
// Type: Creature — Beholder
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Antimagic Cone — Each opponent sacrifices an enchantment of their choice.
//	• Fear Ray — Creatures you control gain menace until end of turn. (A creature with menace can't be blocked except by two or more creatures.)
var BalefulBeholder = newBalefulBeholder()

func newBalefulBeholder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Baleful Beholder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beholder},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
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
								Text: "Antimagic Cone — Each opponent sacrifices an enchantment of their choice.",
								Sequence: []game.Instruction{
									{
										Primitive: game.SacrificePermanents{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
										},
									},
								},
							},
							game.Mode{
								Text: "Fear Ray — Creatures you control gain menace until end of turn. (A creature with menace can't be blocked except by two or more creatures.)",
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													AddKeywords: []game.Keyword{
														game.Menace,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			• Antimagic Cone — Each opponent sacrifices an enchantment of their choice.
			• Fear Ray — Creatures you control gain menace until end of turn. (A creature with menace can't be blocked except by two or more creatures.)
		`,
		},
	}
}
