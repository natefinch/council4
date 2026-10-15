package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TakeDown is the card definition for Take Down.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Choose one —
//	• Take Down deals 4 damage to target creature with flying.
//	• Take Down deals 1 damage to each creature with flying.
var TakeDown = newTakeDown

func newTakeDown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Take Down",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Take Down deals 4 damage to target creature with flying.",
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
									Amount:    game.Fixed(4),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Take Down deals 1 damage to each creature with flying.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying})),
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
			• Take Down deals 4 damage to target creature with flying.
			• Take Down deals 1 damage to each creature with flying.
		`,
		},
	}
}
