package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WildestDreams is the card definition for Wildest Dreams.
//
// Type: Sorcery
// Cost: {X}{X}{G}
//
// Oracle text:
//
//	Return X target cards from your graveyard to your hand. Exile Wildest Dreams.
var WildestDreams = newWildestDreams()

func newWildestDreams() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wildest Dreams",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.X,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "target cards from your graveyard",
						Allow:        game.TargetAllowCard,
						TargetZone:   zone.Graveyard,
						Selection:    opt.Val(game.Selection{Controller: game.ControllerYou}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 3},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 4},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 5},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 6},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 7},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 8},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 9},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 10},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 11},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 12},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 13},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 14},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 15},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 16},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 17},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 18},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 19},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.Exile{
							SourceSpell: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return X target cards from your graveyard to your hand. Exile Wildest Dreams.
		`,
		},
	}
}
