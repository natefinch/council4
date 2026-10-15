package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PistusStrike is the card definition for Pistus Strike.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Destroy target creature with flying. Its controller gets a poison counter.
var PistusStrike = newPistusStrike

func newPistusStrike() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Pistus Strike",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
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
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.AddPlayerCounter{
							Amount:      game.Fixed(1),
							Player:      game.ObjectControllerReference(game.TargetPermanentReference(0)),
							CounterKind: counter.Poison,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature with flying. Its controller gets a poison counter.
		`,
		},
	}
}
