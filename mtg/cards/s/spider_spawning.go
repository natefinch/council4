package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpiderSpawning is the card definition for Spider Spawning.
//
// Type: Sorcery
// Cost: {4}{G}
//
// Oracle text:
//
//	Create a 1/2 green Spider creature token with reach for each creature card in your graveyard.
//	Flashback {6}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var SpiderSpawning = newSpiderSpawning()

func newSpiderSpawning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Spider Spawning",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(6), cost.B}},
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
							Source: game.TokenDef(spiderSpawningToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/2 green Spider creature token with reach for each creature card in your graveyard.
			Flashback {6}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var spiderSpawningToken = newSpiderSpawningToken()

func newSpiderSpawningToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spider",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
		},
	}
}
