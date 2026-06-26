package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConquerorSPledge is the card definition for Conqueror's Pledge.
//
// Type: Sorcery
// Cost: {2}{W}{W}{W}
//
// Oracle text:
//
//	Kicker {6} (You may pay an additional {6} as you cast this spell.)
//	Create six 1/1 white Kor Soldier creature tokens. If this spell was kicked, create twelve of those tokens instead.
var ConquerorSPledge = newConquerorSPledge()

func newConquerorSPledge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Conqueror's Pledge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(6)}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(6),
							Source: game.TokenDef(conquerorSPledgeToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:         true,
								SpellWasKicked: true,
							}),
						}),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(12),
							Source: game.TokenDef(conquerorSPledgeToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {6} (You may pay an additional {6} as you cast this spell.)
			Create six 1/1 white Kor Soldier creature tokens. If this spell was kicked, create twelve of those tokens instead.
		`,
		},
	}
}

var conquerorSPledgeToken = newConquerorSPledgeToken()

func newConquerorSPledgeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Kor Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
