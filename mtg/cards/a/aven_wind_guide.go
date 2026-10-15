package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AvenWindGuide is the card definition for Aven Wind Guide.
//
// Type: Creature — Bird Warrior
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Flying, vigilance
//	Creature tokens you control have flying and vigilance.
//	Embalm {4}{W}{U} ({4}{W}{U}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Bird Warrior with no mana cost. Embalm only as a sorcery.)
var AvenWindGuide = newAvenWindGuide

func newAvenWindGuide() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Aven Wind Guide",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "Creature tokens you control have flying and vigilance.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying, game.Vigilance},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EmbalmActivatedBody(cost.Mana{cost.O(4), cost.W, cost.U}, types.Sub("Bird"), types.Sub("Warrior")),
			},
			OracleText: `
			Flying, vigilance
			Creature tokens you control have flying and vigilance.
			Embalm {4}{W}{U} ({4}{W}{U}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Bird Warrior with no mana cost. Embalm only as a sorcery.)
		`,
		},
	}
}
