package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BlueSunSZenith is the card definition for Blue Sun's Zenith.
//
// Type: Instant
// Cost: {X}{U}{U}{U}
//
// Oracle text:
//
//	Target player draws X cards. Shuffle Blue Sun's Zenith into its owner's library.
var BlueSunSZenith = newBlueSunSZenith

func newBlueSunSZenith() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Blue Sun's Zenith",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Library,
							Shuffle:     true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player draws X cards. Shuffle Blue Sun's Zenith into its owner's library.
		`,
		},
	}
}
