package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SunderingGrowth is the card definition for Sundering Growth.
//
// Type: Instant
// Cost: {G/W}{G/W}
//
// Oracle text:
//
//	Destroy target artifact or enchantment, then populate. (Create a token that's a copy of a creature token you control.)
var SunderingGrowth = newSunderingGrowth

func newSunderingGrowth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Sundering Growth",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or enchantment",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceChosenControlledCreatureToken,
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target artifact or enchantment, then populate. (Create a token that's a copy of a creature token you control.)
		`,
		},
	}
}
