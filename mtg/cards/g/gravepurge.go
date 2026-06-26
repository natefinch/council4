package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Gravepurge is the card definition for Gravepurge.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Put any number of target creature cards from your graveyard on top of your library.
//	Draw a card.
var Gravepurge = newGravepurge()

func newGravepurge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gravepurge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 99,
						Constraint: "any number of target creature cards from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 3},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 4},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 5},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 6},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 7},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 8},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 9},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 10},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 11},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 12},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 13},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 14},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 15},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 16},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 17},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 18},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 19},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 20},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 21},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 22},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 23},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 24},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 25},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 26},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 27},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 28},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 29},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 30},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 31},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 32},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 33},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 34},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 35},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 36},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 37},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 38},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 39},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 40},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 41},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 42},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 43},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 44},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 45},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 46},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 47},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 48},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 49},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 50},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 51},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 52},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 53},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 54},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 55},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 56},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 57},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 58},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 59},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 60},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 61},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 62},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 63},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 64},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 65},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 66},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 67},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 68},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 69},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 70},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 71},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 72},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 73},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 74},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 75},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 76},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 77},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 78},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 79},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 80},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 81},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 82},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 83},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 84},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 85},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 86},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 87},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 88},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 89},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 90},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 91},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 92},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 93},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 94},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 95},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 96},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 97},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 98},
							FromZone:    zone.Graveyard,
							Destination: zone.Library,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put any number of target creature cards from your graveyard on top of your library.
			Draw a card.
		`,
		},
	}
}
