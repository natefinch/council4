package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WindsOfRebuke is the card definition for Winds of Rebuke.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Return target nonland permanent to its owner's hand. Each player mills two cards.
var WindsOfRebuke = newWindsOfRebuke

func newWindsOfRebuke() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Winds of Rebuke",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(2),
							PlayerGroup: game.AllPlayersReference(),
							Destination: zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target nonland permanent to its owner's hand. Each player mills two cards.
		`,
		},
	}
}
