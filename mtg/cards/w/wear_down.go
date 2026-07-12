package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WearDown is the card definition for Wear Down.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Gift a card (You may promise an opponent a gift as you cast this spell. If you do, they draw a card before its other effects.)
//	Destroy target artifact or enchantment. If the gift was promised, instead destroy two target artifacts and/or enchantments.
var WearDown = newWearDown

func newWearDown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wear Down",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
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
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or enchantment",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
						Gate:       game.TargetGateGiftNotPromised,
					},
					game.TargetSpec{
						MinTargets: 2,
						MaxTargets: 2,
						Constraint: "two target artifacts and/or enchantments",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
						Gate:       game.TargetGateGiftPromised,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								GiftPromised: true,
							}),
						}),
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(1),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								GiftPromised: true,
							}),
						}),
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(2),
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
			Destroy target artifact or enchantment. If the gift was promised, instead destroy two target artifacts and/or enchantments.
		`,
		},
	}
}
