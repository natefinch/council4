package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Skeletonize is the card definition for Skeletonize.
//
// Type: Instant
// Cost: {4}{R}
//
// Oracle text:
//
//	Skeletonize deals 3 damage to target creature. When a creature dealt damage this way dies this turn, create a 1/1 black Skeleton creature token with "{B}: Regenerate this token."
var Skeletonize = newSkeletonize

func newSkeletonize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Skeletonize",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(skeletonizeToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Skeletonize deals 3 damage to target creature. When a creature dealt damage this way dies this turn, create a 1/1 black Skeleton creature token with "{B}: Regenerate this token."
		`,
		},
	}
}

var skeletonizeToken = newSkeletonizeToken()

func newSkeletonizeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Skeleton",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Skeleton},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{B}: Regenerate this token.",
					ManaCost:       opt.Val(cost.Mana{cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
