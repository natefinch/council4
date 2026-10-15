package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurnToDust is the card definition for Turn to Dust.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Destroy target Equipment. Add {G}.
var TurnToDust = newTurnToDust

func newTurnToDust() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Turn to Dust",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target Equipment",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Equipment")}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.G,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target Equipment. Add {G}.
		`,
		},
	}
}
