package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CephalopodSentry is the card definition for Cephalopod Sentry.
//
// Type: Artifact Creature — Phyrexian Squid
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Flying
//	Cephalopod Sentry's power is equal to the number of artifacts you control.
var CephalopodSentry = newCephalopodSentry

func newCephalopodSentry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Cephalopod Sentry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:       []color.Color{color.Blue, color.White},
			Types:        []types.Card{types.Artifact, types.Creature},
			Subtypes:     []types.Sub{types.Phyrexian, types.Squid},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 5}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
			Cephalopod Sentry's power is equal to the number of artifacts you control.
		`,
		},
	}
}
