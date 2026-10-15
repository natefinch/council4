package a

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

// AzoriusLocket is the card definition for Azorius Locket.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {W} or {U}.
//	{W/U}{W/U}{W/U}{W/U}, {T}, Sacrifice this artifact: Draw two cards.
var AzoriusLocket = newAzoriusLocket

func newAzoriusLocket() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Azorius Locket",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W/U}{W/U}{W/U}{W/U}, {T}, Sacrifice this artifact: Draw two cards.",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.W, mana.U),
						cost.HybridMana(mana.W, mana.U),
						cost.HybridMana(mana.W, mana.U),
						cost.HybridMana(mana.W, mana.U),
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
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			OracleText: `
			{T}: Add {W} or {U}.
			{W/U}{W/U}{W/U}{W/U}, {T}, Sacrifice this artifact: Draw two cards.
		`,
		},
	}
}
