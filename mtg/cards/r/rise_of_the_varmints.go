package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RiseOfTheVarmints is the card definition for Rise of the Varmints.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Create X 2/1 green Varmint creature tokens, where X is the number of creature cards in your graveyard.
//	Plot {2}{G} (You may pay {2}{G} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
var RiseOfTheVarmints = newRiseOfTheVarmints

func newRiseOfTheVarmints() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rise of the Varmints",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
							}),
							Source: game.TokenDef(riseOfTheVarmintsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 2/1 green Varmint creature tokens, where X is the number of creature cards in your graveyard.
			Plot {2}{G} (You may pay {2}{G} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
		`,
		},
	}
}

var riseOfTheVarmintsToken = newRiseOfTheVarmintsToken()

func newRiseOfTheVarmintsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Varmint",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Varmint},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
