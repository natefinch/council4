package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DenyReality is the card definition for Deny Reality.
//
// Type: Sorcery
// Cost: {3}{U}{B}
//
// Oracle text:
//
//	Cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order.)
//	Return target permanent to its owner's hand.
var DenyReality = newDenyReality

func newDenyReality() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Deny Reality",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.CascadeStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent",
						Allow:      game.TargetAllowPermanent,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order.)
			Return target permanent to its owner's hand.
		`,
		},
	}
}
