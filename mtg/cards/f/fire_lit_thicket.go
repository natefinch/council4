package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FireLitThicket is the card definition for Fire-Lit Thicket.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.
var FireLitThicket = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Red),
		CardFace: game.CardFace{
			Name:  "Fire-Lit Thicket",
			Types: []types.Card{types.Land},
			OracleText: `
			{T}: Add {C}.
			{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.
		`,
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				{
					Text: `
					{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.
				`,
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.R, mana.G),
					}),
					AdditionalCosts: cost.Tap,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Add {R}{R}.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.R,
										},
									},
									{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.R,
										},
									},
								},
							},
							game.Mode{
								Text: "Add {R}{G}.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.R,
										},
									},
									{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
									},
								},
							},
							game.Mode{
								Text: "Add {G}{G}.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
									},
									{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
