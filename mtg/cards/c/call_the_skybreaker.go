package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CallTheSkybreaker is the card definition for Call the Skybreaker.
//
// Type: Sorcery
// Cost: {5}{U/R}{U/R}
//
// Oracle text:
//
//	Create a 5/5 blue and red Elemental creature token with flying.
//	Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
var CallTheSkybreaker = newCallTheSkybreaker

func newCallTheSkybreaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Call the Skybreaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.HybridMana(mana.U, mana.R),
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.RetraceStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(callTheSkybreakerToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 5/5 blue and red Elemental creature token with flying.
			Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
		`,
		},
	}
}

var callTheSkybreakerToken = newCallTheSkybreakerToken()

func newCallTheSkybreakerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Blue, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
