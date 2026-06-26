package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IncreasingDevotion is the card definition for Increasing Devotion.
//
// Type: Sorcery
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Create five 1/1 white Human creature tokens. If this spell was cast from a graveyard, create ten of those tokens instead.
//	Flashback {7}{W}{W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var IncreasingDevotion = newIncreasingDevotion()

func newIncreasingDevotion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Increasing Devotion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(7), cost.W, cost.W}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(5),
							Source: game.TokenDef(increasingDevotionToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								CastFromZone: opt.Val(zone.Graveyard),
							}),
						}),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(10),
							Source: game.TokenDef(increasingDevotionToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								CastFromZone: opt.Val(zone.Graveyard),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Create five 1/1 white Human creature tokens. If this spell was cast from a graveyard, create ten of those tokens instead.
			Flashback {7}{W}{W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var increasingDevotionToken = newIncreasingDevotionToken()

func newIncreasingDevotionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
