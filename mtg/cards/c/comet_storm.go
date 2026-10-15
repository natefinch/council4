package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CometStorm is the card definition for Comet Storm.
//
// Type: Instant
// Cost: {X}{R}{R}
//
// Oracle text:
//
//	Multikicker {1} (You may pay an additional {1} any number of times as you cast this spell.)
//	Choose any target, then choose another target for each time this spell was kicked. Comet Storm deals X damage to each of them.
var CometStorm = newCometStorm

func newCometStorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Comet Storm",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multikicker {1}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(1),
							},
							Multi: true,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:               1,
						MaxTargets:               21,
						Constraint:               "any target",
						Allow:                    game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
						CountEqualsKickerPlusOne: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient:  game.AnyTargetDamageRecipient(0),
							EachTarget: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Multikicker {1} (You may pay an additional {1} any number of times as you cast this spell.)
			Choose any target, then choose another target for each time this spell was kicked. Comet Storm deals X damage to each of them.
		`,
		},
	}
}
