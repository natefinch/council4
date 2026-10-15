package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TumbleweedRising is the card definition for Tumbleweed Rising.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Create an X/X green Elemental creature token, where X is the greatest power among creatures you control.
//	Plot {2}{G} (You may pay {2}{G} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
var TumbleweedRising = newTumbleweedRising

func newTumbleweedRising() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tumbleweed Rising",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Plot {2}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.PlotKeyword{
							Cost: cost.Mana{
								cost.O(2),
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
							Source: game.TokenDef(tumbleweedRisingToken),
							Power: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestPowerInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							})),
							Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestPowerInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create an X/X green Elemental creature token, where X is the greatest power among creatures you control.
			Plot {2}{G} (You may pay {2}{G} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
		`,
		},
	}
}

var tumbleweedRisingToken = newTumbleweedRisingToken()

func newTumbleweedRisingToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Elemental",
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Elemental},
		},
	}
}
