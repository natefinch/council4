package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MayaelTheAnima is the card definition for Mayael the Anima.
//
// Type: Legendary Creature — Elf Shaman
// Cost: {R}{G}{W}
//
// Oracle text:
//
//	{3}{R}{G}{W}, {T}: Look at the top five cards of your library. You may put a creature card with power 5 or greater from among them onto the battlefield. Put the rest on the bottom of your library in any order.
var MayaelTheAnima = newMayaelTheAnima()

func newMayaelTheAnima() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Mayael the Anima",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Shaman},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{3}{R}{G}{W}, {T}: Look at the top five cards of your library. You may put a creature card with power 5 or greater from among them onto the battlefield. Put the rest on the bottom of your library in any order.",
					ManaCost:        opt.Val(cost.Mana{cost.O(3), cost.R, cost.G, cost.W}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:      game.ControllerReference(),
									Look:        game.Fixed(5),
									Take:        game.Fixed(1),
									Remainder:   game.DigRemainderLibraryBottom,
									Filter:      opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5})}),
									TakeUpTo:    true,
									Destination: zone.Battlefield,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{3}{R}{G}{W}, {T}: Look at the top five cards of your library. You may put a creature card with power 5 or greater from among them onto the battlefield. Put the rest on the bottom of your library in any order.
		`,
		},
	}
}
