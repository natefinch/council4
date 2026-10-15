package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HungryForMore is the card definition for Hungry for More.
//
// Type: Sorcery
// Cost: {B}{R}
//
// Oracle text:
//
//	Create a 3/1 black and red Vampire creature token with trample, lifelink, and haste. Sacrifice it at the beginning of the next end step.
//	Flashback {1}{B}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var HungryForMore = newHungryForMore

func newHungryForMore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Hungry for More",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {1}{B}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.B,
								cost.R,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(1),
							Source:        game.TokenDef(hungryForMoreToken),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Sacrifice{
														Group: game.CapturedObjectsGroup(),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedObjectGroup: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 3/1 black and red Vampire creature token with trample, lifelink, and haste. Sacrifice it at the beginning of the next end step.
			Flashback {1}{B}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var hungryForMoreToken = newHungryForMoreToken()

func newHungryForMoreToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Vampire",
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.LifelinkStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
