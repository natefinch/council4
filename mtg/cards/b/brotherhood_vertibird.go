package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrotherhoodVertibird is the card definition for Brotherhood Vertibird.
//
// Type: Artifact — Vehicle
// Cost: {3}
//
// Oracle text:
//
//	Flying
//	Brotherhood Vertibird's power is equal to the number of artifacts you control.
//	Crew 2 (Tap any number of creatures you control with total power 2 or greater: This Vehicle becomes an artifact creature until end of turn.)
var BrotherhoodVertibird = newBrotherhoodVertibird

func newBrotherhoodVertibird() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Brotherhood Vertibird",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:        []types.Card{types.Artifact},
			Subtypes:     []types.Sub{types.Vehicle},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(2),
			},
			OracleText: `
			Flying
			Brotherhood Vertibird's power is equal to the number of artifacts you control.
			Crew 2 (Tap any number of creatures you control with total power 2 or greater: This Vehicle becomes an artifact creature until end of turn.)
		`,
		},
	}
}
