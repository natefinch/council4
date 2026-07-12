package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ClutchOfTheUndercity is the card definition for Clutch of the Undercity.
//
// Type: Instant
// Cost: {1}{U}{U}{B}
//
// Oracle text:
//
//	Return target permanent to its owner's hand. Its controller loses 3 life.
//	Transmute {1}{U}{B} ({1}{U}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
var ClutchOfTheUndercity = newClutchOfTheUndercity

func newClutchOfTheUndercity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Clutch of the Undercity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.TransmuteActivatedAbility(cost.Mana{cost.O(1), cost.U, cost.B}, 4),
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
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(3),
							Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target permanent to its owner's hand. Its controller loses 3 life.
			Transmute {1}{U}{B} ({1}{U}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
		`,
		},
	}
}
