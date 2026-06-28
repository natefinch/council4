package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MaelstromPulse is the card definition for Maelstrom Pulse.
//
// Type: Sorcery
// Cost: {1}{B}{G}
//
// Oracle text:
//
//	Destroy target nonland permanent and all other permanents with the same name as that permanent.
var MaelstromPulse = newMaelstromPulse()

func newMaelstromPulse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Maelstrom Pulse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent and all other permanents with the same name as that permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target nonland permanent and all other permanents with the same name as that permanent.
		`,
		},
	}
}
