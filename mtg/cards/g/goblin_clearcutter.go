package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinClearcutter is the card definition for Goblin Clearcutter.
//
// Type: Creature — Goblin
// Cost: {3}{R}
//
// Oracle text:
//
//	{T}, Sacrifice a Forest: Add three mana in any combination of {R} and/or {G}.
var GoblinClearcutter = newGoblinClearcutter

func newGoblinClearcutter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Goblin Clearcutter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
