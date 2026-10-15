package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RevengeOfTheRats is the card definition for Revenge of the Rats.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Create a tapped 1/1 black Rat creature token for each creature card in your graveyard.
//	Flashback {2}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var RevengeOfTheRats = newRevengeOfTheRats

func newRevengeOfTheRats() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Revenge of the Rats",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {2}{B}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.B,
								cost.B,
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
							Source:      game.TokenDef(revengeOfTheRatsToken),
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a tapped 1/1 black Rat creature token for each creature card in your graveyard.
			Flashback {2}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var revengeOfTheRatsToken = newRevengeOfTheRatsToken()

func newRevengeOfTheRatsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rat",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
