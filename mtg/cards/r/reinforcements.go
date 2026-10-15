package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Reinforcements is the card definition for Reinforcements.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Put up to three target creature cards from your graveyard on top of your library.
var Reinforcements = newReinforcements

func newReinforcements() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Reinforcements",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 3,
						Constraint: "up to three target creature cards from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
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
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind:        game.CardReferenceTarget,
								TargetIndex: 1,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind:        game.CardReferenceTarget,
								TargetIndex: 2,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put up to three target creature cards from your graveyard on top of your library.
		`,
		},
	}
}
