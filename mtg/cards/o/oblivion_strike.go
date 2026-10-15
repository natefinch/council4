package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OblivionStrike is the card definition for Oblivion Strike.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Exile target creature.
var OblivionStrike = newOblivionStrike

func newOblivionStrike() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Oblivion Strike",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Types: []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
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
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Devoid (This card has no color.)
			Exile target creature.
		`,
		},
	}
}
