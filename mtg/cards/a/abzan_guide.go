package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbzanGuide is the card definition for Abzan Guide.
//
// Type: Creature — Human Warrior
// Cost: {3}{W}{B}{G}
//
// Oracle text:
//
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	Morph {2}{W}{B}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var AbzanGuide = newAbzanGuide

func newAbzanGuide() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Abzan Guide",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Morph {2}{W}{B}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.W,
								cost.B,
								cost.G,
							},
						},
					},
				},
			},
			OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			Morph {2}{W}{B}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
