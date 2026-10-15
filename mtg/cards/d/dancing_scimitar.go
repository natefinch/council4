package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DancingScimitar is the card definition for Dancing Scimitar.
//
// Type: Artifact Creature — Spirit
// Cost: {4}
//
// Oracle text:
//
//	Flying (This creature can't be blocked except by creatures with flying or reach.)
var DancingScimitar = newDancingScimitar

func newDancingScimitar() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dancing Scimitar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying (This creature can't be blocked except by creatures with flying or reach.)
		`,
		},
	}
}
