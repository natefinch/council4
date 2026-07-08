package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VampireHexmage is the card definition for Vampire Hexmage.
//
// Type: Creature — Vampire Shaman
// Cost: {B}{B}
//
// Oracle text:
//
//	First strike
//	Sacrifice this creature: Remove all counters from target permanent.
var VampireHexmage = newVampireHexmage

func newVampireHexmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vampire Hexmage",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice this creature: Remove all counters from target permanent.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent",
								Allow:      game.TargetAllowPermanent,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Object:   game.TargetPermanentReference(0),
									AllKinds: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			First strike
			Sacrifice this creature: Remove all counters from target permanent.
		`,
		},
	}
}
