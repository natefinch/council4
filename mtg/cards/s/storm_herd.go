package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StormHerd is the card definition for Storm Herd.
//
// Type: Sorcery
// Cost: {8}{W}{W}
//
// Oracle text:
//
//	Create X 1/1 white Pegasus creature tokens with flying, where X is your life total.
var StormHerd = newStormHerd

func newStormHerd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Storm Herd",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountControllerLife,
								Multiplier: 1,
							}),
							Source: game.TokenDef(stormHerdToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 white Pegasus creature tokens with flying, where X is your life total.
		`,
		},
	}
}

var stormHerdToken = newStormHerdToken()

func newStormHerdToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Pegasus",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pegasus},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
