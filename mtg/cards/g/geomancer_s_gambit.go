package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GeomancerSGambit is the card definition for Geomancer's Gambit.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Destroy target land. Its controller may search their library for a basic land card, put it onto the battlefield, then shuffle.
//	Draw a card.
var GeomancerSGambit = newGeomancerSGambit

func newGeomancerSGambit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Geomancer's Gambit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target land",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Search{
							Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Battlefield,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
							},
							Amount: game.Fixed(1),
						},
						Optional:      true,
						OptionalActor: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
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
			Destroy target land. Its controller may search their library for a basic land card, put it onto the battlefield, then shuffle.
			Draw a card.
		`,
		},
	}
}
