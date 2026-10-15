package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SummonBahamut is the card definition for Summon: Bahamut.
//
// Type: Enchantment Creature — Saga Dragon
// Cost: {9}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
//	I, II — Destroy up to one target nonland permanent.
//	III — Draw two cards.
//	IV — Mega Flare — This creature deals damage equal to the total mana value of other permanents you control to each opponent.
//	Flying
var SummonBahamut = newSummonBahamut

func newSummonBahamut() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Summon: Bahamut",
			ManaCost: opt.Val(cost.Mana{
				cost.O(9),
			}),
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Saga, types.Dragon},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II — Destroy up to one target nonland permanent.",
					Chapters: []int{1, 2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 1,
										Constraint: "up to one target nonland permanent",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III — Draw two cards.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "IV — Mega Flare — This creature deals damage equal to the total mana value of other permanents you control to each opponent.",
					Chapters: []int{4},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountTotalManaValueInGroup,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou, ExcludeSource: true}),
											}),
											Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
			I, II — Destroy up to one target nonland permanent.
			III — Draw two cards.
			IV — Mega Flare — This creature deals damage equal to the total mana value of other permanents you control to each opponent.
			Flying
		`,
		},
	}
}
