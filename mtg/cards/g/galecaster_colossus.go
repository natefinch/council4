package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GalecasterColossus is the card definition for Galecaster Colossus.
//
// Type: Creature — Giant Wizard
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Tap an untapped Wizard you control: Return target nonland permanent you don't control to its owner's hand.
var GalecasterColossus = newGalecasterColossus

func newGalecasterColossus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Galecaster Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Wizard},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap an untapped Wizard you control: Return target nonland permanent you don't control to its owner's hand.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Wizard you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{0: types.Wizard},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target nonland permanent you don't control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerNotYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Hand,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Tap an untapped Wizard you control: Return target nonland permanent you don't control to its owner's hand.
		`,
		},
	}
}
