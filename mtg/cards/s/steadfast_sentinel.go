package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SteadfastSentinel is the card definition for Steadfast Sentinel.
//
// Type: Creature — Human Cleric
// Cost: {3}{W}
//
// Oracle text:
//
//	Vigilance
//	Eternalize {4}{W}{W} ({4}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Human Cleric with no mana cost. Eternalize only as a sorcery.)
var SteadfastSentinel = newSteadfastSentinel

func newSteadfastSentinel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Steadfast Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EternalizeActivatedBody(cost.Mana{cost.O(4), cost.W, cost.W}, types.Sub("Human"), types.Sub("Cleric")),
			},
			OracleText: `
			Vigilance
			Eternalize {4}{W}{W} ({4}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Human Cleric with no mana cost. Eternalize only as a sorcery.)
		`,
		},
	}
}
