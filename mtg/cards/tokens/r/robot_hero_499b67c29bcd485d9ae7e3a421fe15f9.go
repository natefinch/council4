package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RobotHero is the card definition for Robot Hero.
//
// Type: Token Artifact Creature — Robot Hero
//
// Oracle text:
//   Flying (This token can't be blocked except by creatures with flying or reach.)

// RobotHeroToken499b67c29bcd485d9ae7e3a421fe15f9 is the card definition for Robot Hero.
var RobotHeroToken499b67c29bcd485d9ae7e3a421fe15f9 = newRobotHeroToken499b67c29bcd485d9ae7e3a421fe15f9()

func newRobotHeroToken499b67c29bcd485d9ae7e3a421fe15f9() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Robot Hero",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot, types.Hero},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying (This token can't be blocked except by creatures with flying or reach.)
		`,
		},
	}
}
