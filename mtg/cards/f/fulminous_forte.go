package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FulminousForte is the card definition for Fulminous Forte.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Choose one —
//	• Fulminous Forte deals 1 damage to each creature and planeswalker your opponents control.
//	• Fulminous Forte deals 5 damage to target creature or planeswalker.
var FulminousForte = newFulminousForte

func newFulminousForte() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fulminous Forte",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Fulminous Forte deals 1 damage to each creature and planeswalker your opponents control.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerOpponent})),
								},
							},
						},
					},
					game.Mode{
						Text: "Fulminous Forte deals 5 damage to target creature or planeswalker.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature or planeswalker",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(5),
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
			• Fulminous Forte deals 1 damage to each creature and planeswalker your opponents control.
			• Fulminous Forte deals 5 damage to target creature or planeswalker.
		`,
		},
	}
}
