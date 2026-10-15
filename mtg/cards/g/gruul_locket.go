package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GruulLocket is the card definition for Gruul Locket.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {R} or {G}.
//	{R/G}{R/G}{R/G}{R/G}, {T}, Sacrifice this artifact: Draw two cards.
var GruulLocket = newGruulLocket

func newGruulLocket() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Gruul Locket",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{R/G}{R/G}{R/G}{R/G}, {T}, Sacrifice this artifact: Draw two cards.",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.R, mana.G),
						cost.HybridMana(mana.R, mana.G),
						cost.HybridMana(mana.R, mana.G),
						cost.HybridMana(mana.R, mana.G),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			OracleText: `
			{T}: Add {R} or {G}.
			{R/G}{R/G}{R/G}{R/G}, {T}, Sacrifice this artifact: Draw two cards.
		`,
		},
	}
}
