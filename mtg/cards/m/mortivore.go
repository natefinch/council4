package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Mortivore is the card definition for Mortivore.
//
// Type: Creature — Lhurgoyf
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Mortivore's power and toughness are each equal to the number of creature cards in all graveyards.
//	{B}: Regenerate this creature. (The next time this creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
var Mortivore = newMortivore

func newMortivore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mortivore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:           []color.Color{color.Black},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Lhurgoyf},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueCreatureCardsInAllGraveyards}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueCreatureCardsInAllGraveyards}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}: Regenerate this creature. (The next time this creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Regenerate{
											Object: game.SourcePermanentReference(),
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
			Mortivore's power and toughness are each equal to the number of creature cards in all graveyards.
			{B}: Regenerate this creature. (The next time this creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
		`,
		},
	}
}
