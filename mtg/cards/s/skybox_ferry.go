package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkyboxFerry is the card definition for Skybox Ferry.
//
// Type: Artifact — Vehicle
// Cost: {5}
//
// Oracle text:
//
//	Flying
//	Crew 2 (Tap any number of creatures you control with total power 2 or more: This Vehicle becomes an artifact creature until end of turn.)
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var SkyboxFerry = newSkyboxFerry

func newSkyboxFerry() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Skybox Ferry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(2),
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Flying
			Crew 2 (Tap any number of creatures you control with total power 2 or more: This Vehicle becomes an artifact creature until end of turn.)
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
