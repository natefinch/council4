package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BirdSoldier is the card definition for Bird Soldier.
//
// Type: Token Creature — Bird Soldier
//
// Oracle text:
//   Flying

// BirdSoldierToken9e44ccd6134d483fa98c222012e30af5 is the card definition for Bird Soldier.
var BirdSoldierToken9e44ccd6134d483fa98c222012e30af5 = newBirdSoldierToken9e44ccd6134d483fa98c222012e30af5()

func newBirdSoldierToken9e44ccd6134d483fa98c222012e30af5() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird Soldier",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
		`,
		},
	}
}
