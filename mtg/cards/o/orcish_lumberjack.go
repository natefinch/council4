package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrcishLumberjack is the card definition for Orcish Lumberjack.
//
// Type: Creature — Orc
// Cost: {R}
//
// Oracle text:
//
//	{T}, Sacrifice a Forest: Add three mana in any combination of {R} and/or {G}.
var OrcishLumberjack = newOrcishLumberjack

func newOrcishLumberjack() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Orcish Lumberjack",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Forest",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Forest},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:            game.Fixed(3),
									CombinationColors: []mana.Color{mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}, Sacrifice a Forest: Add three mana in any combination of {R} and/or {G}.
		`,
		},
	}
}
