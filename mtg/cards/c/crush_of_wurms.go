package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrushOfWurms is the card definition for Crush of Wurms.
//
// Type: Sorcery
// Cost: {6}{G}{G}{G}
//
// Oracle text:
//
//	Create three 6/6 green Wurm creature tokens.
//	Flashback {9}{G}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var CrushOfWurms = newCrushOfWurms

func newCrushOfWurms() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Crush of Wurms",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {9}{G}{G}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(9),
								cost.G,
								cost.G,
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
							Amount: game.Fixed(3),
							Source: game.TokenDef(crushOfWurmsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 6/6 green Wurm creature tokens.
			Flashback {9}{G}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var crushOfWurmsToken = newCrushOfWurmsToken()

func newCrushOfWurmsToken() *game.CardDef {
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
