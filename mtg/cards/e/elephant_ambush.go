package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElephantAmbush is the card definition for Elephant Ambush.
//
// Type: Instant
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Create a 3/3 green Elephant creature token.
//	Flashback {6}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ElephantAmbush = newElephantAmbush

func newElephantAmbush() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Elephant Ambush",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
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
							Amount: game.Fixed(1),
							Source: game.TokenDef(elephantAmbushToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 3/3 green Elephant creature token.
			Flashback {6}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var elephantAmbushToken = newElephantAmbushToken()

func newElephantAmbushToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elephant",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
