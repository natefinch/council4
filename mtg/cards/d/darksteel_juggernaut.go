package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelJuggernaut is the card definition for Darksteel Juggernaut.
//
// Type: Artifact Creature — Juggernaut
// Cost: {5}
//
// Oracle text:
//
//	Indestructible
//	Darksteel Juggernaut's power and toughness are each equal to the number of artifacts you control.
//	This creature attacks each combat if able.
var DarksteelJuggernaut = newDarksteelJuggernaut

func newDarksteelJuggernaut() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Juggernaut",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:            []types.Card{types.Artifact, types.Creature},
			Subtypes:         []types.Sub{types.Juggernaut},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Indestructible
			Darksteel Juggernaut's power and toughness are each equal to the number of artifacts you control.
			This creature attacks each combat if able.
		`,
		},
	}
}
