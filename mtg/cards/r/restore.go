package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Restore is the card definition for Restore.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Put target land card from a graveyard onto the battlefield under your control.
var Restore = newRestore

func newRestore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Restore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target land card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							Recipient: opt.Val(game.ControllerReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put target land card from a graveyard onto the battlefield under your control.
		`,
		},
	}
}
