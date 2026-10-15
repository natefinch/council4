package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TrollsOfTelJilad is the card definition for Trolls of Tel-Jilad.
//
// Type: Creature — Troll Shaman
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	{1}{G}: Regenerate target green creature.
var TrollsOfTelJilad = newTrollsOfTelJilad

func newTrollsOfTelJilad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Trolls of Tel-Jilad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Troll, types.Shaman},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{G}: Regenerate target green creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target green creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Regenerate{
											Object: game.TargetPermanentReference(0),
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
			{1}{G}: Regenerate target green creature.
		`,
		},
	}
}
