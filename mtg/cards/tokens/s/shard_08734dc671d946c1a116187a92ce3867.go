package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Shard is the card definition for Shard.
//
// Type: Token Enchantment — Shard
//
// Oracle text:
//   {2}, Sacrifice this enchantment: Scry 1, then draw a card.

// ShardToken08734dc671d946c1a116187a92ce3867 is the card definition for Shard.
var ShardToken08734dc671d946c1a116187a92ce3867 = newShardToken08734dc671d946c1a116187a92ce3867()

func newShardToken08734dc671d946c1a116187a92ce3867() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Shard",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Shard},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, Sacrifice this enchantment: Scry 1, then draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this enchantment",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Scry{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
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
			},
			OracleText: `
			{2}, Sacrifice this enchantment: Scry 1, then draw a card.
		`,
		},
	}
}
