package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoanOfTheUnhallowed is the card definition for Moan of the Unhallowed.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Create two 2/2 black Zombie creature tokens.
//	Flashback {5}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var MoanOfTheUnhallowed = newMoanOfTheUnhallowed

func newMoanOfTheUnhallowed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Moan of the Unhallowed",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {5}{B}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(5),
								cost.B,
								cost.B,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(moanOfTheUnhallowedToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 2/2 black Zombie creature tokens.
			Flashback {5}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var moanOfTheUnhallowedToken = newMoanOfTheUnhallowedToken()

func newMoanOfTheUnhallowedToken() *game.CardDef {
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
