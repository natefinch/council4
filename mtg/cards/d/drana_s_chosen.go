package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DranaSChosen is the card definition for Drana's Chosen.
//
// Type: Creature — Vampire Shaman Ally
// Cost: {3}{B}
//
// Oracle text:
//
//	Cohort — {T}, Tap an untapped Ally you control: Create a tapped 2/2 black Zombie creature token.
var DranaSChosen = newDranaSChosen

func newDranaSChosen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Drana's Chosen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Shaman, types.Ally},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Cohort — {T}, Tap an untapped Ally you control: Create a tapped 2/2 black Zombie creature token.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Ally you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Ally},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:      game.Fixed(1),
									Source:      game.TokenDef(dranaSChosenToken),
									EntryTapped: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Cohort — {T}, Tap an untapped Ally you control: Create a tapped 2/2 black Zombie creature token.
		`,
		},
	}
}

var dranaSChosenToken = newDranaSChosenToken()

func newDranaSChosenToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
