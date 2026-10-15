package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoarOfTheWurm is the card definition for Roar of the Wurm.
//
// Type: Sorcery
// Cost: {6}{G}
//
// Oracle text:
//
//	Create a 6/6 green Wurm creature token.
//	Flashback {3}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var RoarOfTheWurm = newRoarOfTheWurm

func newRoarOfTheWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Roar of the Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {3}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.G,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(roarOfTheWurmToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 6/6 green Wurm creature token.
			Flashback {3}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var roarOfTheWurmToken = newRoarOfTheWurmToken()

func newRoarOfTheWurmToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wurm",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
		},
	}
}
