package s

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

// SimicLocket is the card definition for Simic Locket.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {G} or {U}.
//	{G/U}{G/U}{G/U}{G/U}, {T}, Sacrifice this artifact: Draw two cards.
var SimicLocket = newSimicLocket

func newSimicLocket() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Simic Locket",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{G/U}{G/U}{G/U}{G/U}, {T}, Sacrifice this artifact: Draw two cards.",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.G, mana.U),
						cost.HybridMana(mana.G, mana.U),
						cost.HybridMana(mana.G, mana.U),
						cost.HybridMana(mana.G, mana.U),
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
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			OracleText: `
			{T}: Add {G} or {U}.
			{G/U}{G/U}{G/U}{G/U}, {T}, Sacrifice this artifact: Draw two cards.
		`,
		},
	}
}
