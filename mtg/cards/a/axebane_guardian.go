package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AxebaneGuardian is the card definition for Axebane Guardian.
//
// Type: Creature — Human Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	Defender
//	{T}: Add X mana in any combination of colors, where X is the number of creatures you control with defender.
var AxebaneGuardian = newAxebaneGuardian

func newAxebaneGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Axebane Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Keyword: game.Defender}),
									}),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			{T}: Add X mana in any combination of colors, where X is the number of creatures you control with defender.
		`,
		},
	}
}
