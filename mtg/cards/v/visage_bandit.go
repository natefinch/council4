package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VisageBandit is the card definition for Visage Bandit.
//
// Type: Creature — Shapeshifter Rogue
// Cost: {3}{U}
//
// Oracle text:
//
//	You may have this creature enter as a copy of a creature you control, except it's a Shapeshifter Rogue in addition to its other types.
//	Plot {2}{U} (You may pay {2}{U} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
var VisageBandit = newVisageBandit

func newVisageBandit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Visage Bandit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Plot {2}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.PlotKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.U,
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of a creature you control, except it's a Shapeshifter Rogue in addition to its other types.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, true, false, nil, false, nil, []types.Sub{types.Shapeshifter, types.Rogue}),
			},
			OracleText: `
			You may have this creature enter as a copy of a creature you control, except it's a Shapeshifter Rogue in addition to its other types.
			Plot {2}{U} (You may pay {2}{U} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
		`,
		},
	}
}
