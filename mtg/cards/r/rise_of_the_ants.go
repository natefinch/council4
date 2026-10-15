package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiseOfTheAnts is the card definition for Rise of the Ants.
//
// Type: Sorcery
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Create two 3/3 green Insect creature tokens. You gain 2 life.
//	Flashback {6}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var RiseOfTheAnts = newRiseOfTheAnts

func newRiseOfTheAnts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rise of the Ants",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {6}{G}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(6),
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
							Amount: game.Fixed(2),
							Source: game.TokenDef(riseOfTheAntsToken),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 3/3 green Insect creature tokens. You gain 2 life.
			Flashback {6}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var riseOfTheAntsToken = newRiseOfTheAntsToken()

func newRiseOfTheAntsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
