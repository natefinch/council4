package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LiturgyOfBlood is the card definition for Liturgy of Blood.
//
// Type: Sorcery
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Destroy target creature. Add {B}{B}{B}.
var LiturgyOfBlood = newLiturgyOfBlood

func newLiturgyOfBlood() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Liturgy of Blood",
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
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
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
							ManaColor: mana.B,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature. Add {B}{B}{B}.
		`,
		},
	}
}
