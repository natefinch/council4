package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FanaticOfRhonas is the card definition for Fanatic of Rhonas.
//
// Type: Token Creature — Zombie Snake Druid
//
// Oracle text:
//   {T}: Add {G}.
//   Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.

// FanaticOfRhonasToken86aef5a2e9754f358fd1b1de187319e6 is the card definition for Fanatic of Rhonas.
var FanaticOfRhonasToken86aef5a2e9754f358fd1b1de187319e6 = newFanaticOfRhonasToken86aef5a2e9754f358fd1b1de187319e6()

func newFanaticOfRhonasToken86aef5a2e9754f358fd1b1de187319e6() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:      "Fanatic of Rhonas",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Snake, types.Druid},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
				game.ManaAbility{
					Text: "Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if you control a creature with power 4 or greater",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 4,
							})},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
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
			{T}: Add {G}.
			Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.
		`,
		},
	}
}
