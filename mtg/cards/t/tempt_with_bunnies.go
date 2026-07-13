package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TemptWithBunnies is the card definition for Tempt with Bunnies.
//
// Type: Sorcery
// Cost: {2}{W}
//
// Oracle text:
//
//	Tempting Offer — Draw a card and create a 1/1 white Rabbit creature token. Then each opponent may draw a card and create a 1/1 white Rabbit creature token. For each opponent who does, you draw a card and you create a 1/1 white Rabbit creature token.
var TemptWithBunnies = newTemptWithBunnies

func newTemptWithBunnies() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Tempt with Bunnies",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Optional:           true,
						OptionalActorGroup: opt.Val(game.OpponentsReference()),
						TemptingOffer:      true,
						TemptingOfferBody: []game.Instruction{{
							Primitive: game.Draw{
								Amount: game.Fixed(1),
								Player: game.GroupOfferMemberReference(),
							},
						},
							{
								Primitive: game.CreateToken{
									Amount:    game.Fixed(1),
									Source:    game.TokenDef(temptWithBunniesToken),
									Recipient: opt.Val(game.GroupOfferMemberReference()),
								},
							}},
					},
				},
			}.Ability()),
			OracleText: `
			Tempting Offer — Draw a card and create a 1/1 white Rabbit creature token. Then each opponent may draw a card and create a 1/1 white Rabbit creature token. For each opponent who does, you draw a card and you create a 1/1 white Rabbit creature token.
		`,
		},
	}
}

var temptWithBunniesToken = newTemptWithBunniesToken()

func newTemptWithBunniesToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rabbit",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rabbit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
