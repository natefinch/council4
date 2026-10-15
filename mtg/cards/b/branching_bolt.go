package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BranchingBolt is the card definition for Branching Bolt.
//
// Type: Instant
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Choose one or both —
//	• Branching Bolt deals 3 damage to target creature with flying.
//	• Branching Bolt deals 3 damage to target creature without flying.
var BranchingBolt = newBranchingBolt

func newBranchingBolt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Branching Bolt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Branching Bolt deals 3 damage to target creature with flying.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature with flying",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Keyword: game.Flying}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Branching Bolt deals 3 damage to target creature without flying.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature without flying",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedKeyword: game.Flying}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Choose one or both —
			• Branching Bolt deals 3 damage to target creature with flying.
			• Branching Bolt deals 3 damage to target creature without flying.
		`,
		},
	}
}
