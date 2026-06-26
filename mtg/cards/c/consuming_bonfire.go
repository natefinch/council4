package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConsumingBonfire is the card definition for Consuming Bonfire.
//
// Type: Kindred Sorcery — Elemental
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Choose one —
//	• Consuming Bonfire deals 4 damage to target non-Elemental creature.
//	• Consuming Bonfire deals 7 damage to target Treefolk creature.
var ConsumingBonfire = newConsumingBonfire()

func newConsumingBonfire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Consuming Bonfire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Elemental},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Consuming Bonfire deals 4 damage to target non-Elemental creature.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target non-Elemental creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Elemental")}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(4),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Consuming Bonfire deals 7 damage to target Treefolk creature.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Treefolk creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Treefolk")}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(7),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Consuming Bonfire deals 4 damage to target non-Elemental creature.
			• Consuming Bonfire deals 7 damage to target Treefolk creature.
		`,
		},
	}
}
