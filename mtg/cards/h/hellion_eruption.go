package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HellionEruption is the card definition for Hellion Eruption.
//
// Type: Sorcery
// Cost: {5}{R}
//
// Oracle text:
//
//	Sacrifice all creatures you control, then create that many 4/4 red Hellion creature tokens.
var HellionEruption = newHellionEruption

func newHellionEruption() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hellion Eruption",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							All:       true,
							Player:    game.ControllerReference(),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
						PublishResult: game.ResultKey("sacrificed-this-way"),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountPreviousEffectResult,
								ResultKey: game.ResultKey("sacrificed-this-way"),
							}),
							Source: game.TokenDef(hellionEruptionToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Sacrifice all creatures you control, then create that many 4/4 red Hellion creature tokens.
		`,
		},
	}
}

var hellionEruptionToken = newHellionEruptionToken()

func newHellionEruptionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Hellion",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
