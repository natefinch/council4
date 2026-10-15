package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IronGiant is the card definition for Iron Giant.
//
// Type: Artifact Creature — Demon
// Cost: {7}
//
// Oracle text:
//
//	Reach, vigilance, trample
var IronGiant = newIronGiant

func newIronGiant() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Iron Giant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Demon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Reach, vigilance, trample
		`,
		},
	}
}
