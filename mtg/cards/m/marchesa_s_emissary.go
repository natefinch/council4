package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MarchesaSEmissary is the card definition for Marchesa's Emissary.
//
// Type: Creature — Human Rogue
// Cost: {3}{U}
//
// Oracle text:
//
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
//	Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
var MarchesaSEmissary = newMarchesaSEmissary

func newMarchesaSEmissary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Marchesa's Emissary",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.DethroneTriggeredBody,
			},
			OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
			Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
		`,
		},
	}
}
