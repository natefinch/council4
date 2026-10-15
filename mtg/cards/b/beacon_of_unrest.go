package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BeaconOfUnrest is the card definition for Beacon of Unrest.
//
// Type: Sorcery
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Put target artifact or creature card from a graveyard onto the battlefield under your control. Shuffle Beacon of Unrest into its owner's library.
var BeaconOfUnrest = newBeaconOfUnrest

func newBeaconOfUnrest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Beacon of Unrest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or creature card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							Recipient: opt.Val(game.ControllerReference()),
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
			Put target artifact or creature card from a graveyard onto the battlefield under your control. Shuffle Beacon of Unrest into its owner's library.
		`,
		},
	}
}
