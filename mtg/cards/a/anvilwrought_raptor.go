package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnvilwroughtRaptor is the card definition for Anvilwrought Raptor.
//
// Type: Artifact Creature — Bird
// Cost: {4}
//
// Oracle text:
//
//	Flying
//	First strike (This creature deals combat damage before creatures without first strike.)
var AnvilwroughtRaptor = newAnvilwroughtRaptor

func newAnvilwroughtRaptor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Anvilwrought Raptor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Flying
			First strike (This creature deals combat damage before creatures without first strike.)
		`,
		},
	}
}
