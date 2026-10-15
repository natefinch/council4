package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NoxiousRevival is the card definition for Noxious Revival.
//
// Type: Instant
// Cost: {G/P}
//
// Oracle text:
//
//	({G/P} can be paid with either {G} or 2 life.)
//	Put target card from a graveyard on top of its owner's library.
var NoxiousRevival = newNoxiousRevival

func newNoxiousRevival() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Noxious Revival",
			ManaCost: opt.Val(cost.Mana{
				cost.PhyrexianMana(mana.G),
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind: game.CardReferenceTarget,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
				},
			}.Ability()),
			OracleText: `
			({G/P} can be paid with either {G} or 2 life.)
			Put target card from a graveyard on top of its owner's library.
		`,
		},
	}
}
