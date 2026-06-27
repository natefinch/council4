package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CarrionBeetles is the card definition for Carrion Beetles.
//
// Type: Creature — Insect
// Cost: {B}
//
// Oracle text:
//
//	{2}{B}, {T}: Exile up to three target cards from a single graveyard.
var CarrionBeetles = newCarrionBeetles()

func newCarrionBeetles() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Carrion Beetles",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}{B}, {T}: Exile up to three target cards from a single graveyard.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.B}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
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
					}.Ability(),
				},
			},
			OracleText: `
			{2}{B}, {T}: Exile up to three target cards from a single graveyard.
		`,
		},
	}
}
