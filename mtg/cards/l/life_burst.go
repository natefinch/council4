package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LifeBurst is the card definition for Life Burst.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Target player gains 4 life, then gains 4 life for each card named Life Burst in each graveyard.
var LifeBurst = newLifeBurst()

func newLifeBurst() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Life Burst",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(4),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCardsNamedSourceInGraveyards,
								Multiplier: 4,
							}),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player gains 4 life, then gains 4 life for each card named Life Burst in each graveyard.
		`,
		},
	}
}
