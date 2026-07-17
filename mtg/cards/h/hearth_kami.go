package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HearthKami is the card definition for Hearth Kami.
//
// Type: Creature — Spirit
// Cost: {1}{R}
//
// Oracle text:
//
//	{X}, Sacrifice this creature: Destroy target artifact with mana value X.
var HearthKami = newHearthKami

func newHearthKami() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hearth Kami",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{X}, Sacrifice this creature: Destroy target artifact with mana value X.",
					ManaCost: opt.Val(cost.Mana{cost.X}),
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
								MinTargets:       1,
								MaxTargets:       1,
								Constraint:       "target artifact with mana value X",
								Allow:            game.TargetAllowPermanent,
								Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
								ManaValueEqualsX: true,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{X}, Sacrifice this creature: Destroy target artifact with mana value X.
		`,
		},
	}
}
