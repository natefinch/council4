package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpiritOfTheLabyrinth is the card definition for Spirit of the Labyrinth.
//
// Type: Enchantment Creature — Spirit
// Cost: {1}{W}
//
// Oracle text:
//
//	Each player can't draw more than one card each turn.
var SpiritOfTheLabyrinth = newSpiritOfTheLabyrinth

func newSpiritOfTheLabyrinth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Spirit of the Labyrinth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each player can't draw more than one card each turn.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectDrawLimitPerTurn,
							Duration:         game.DurationPermanent,
							DrawLimitPerTurn: 1,
						},
					},
				},
			},
			OracleText: `
			Each player can't draw more than one card each turn.
		`,
		},
	}
}
