package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Decompose is the card definition for Decompose.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Exile up to three target cards from a single graveyard.
var Decompose = newDecompose()

func newDecompose() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Decompose",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:    0,
						MaxTargets:    3,
						Constraint:    "up to three target cards from a single graveyard",
						Allow:         game.TargetAllowCard,
						TargetZone:    zone.Graveyard,
						Selection:     opt.Val(game.Selection{}),
						SameGraveyard: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile up to three target cards from a single graveyard.
		`,
		},
	}
}
