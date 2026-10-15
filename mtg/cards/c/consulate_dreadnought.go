package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConsulateDreadnought is the card definition for Consulate Dreadnought.
//
// Type: Artifact — Vehicle
// Cost: {1}
//
// Oracle text:
//
//	Crew 6 (Tap any number of creatures you control with total power 6 or more: This Vehicle becomes an artifact creature until end of turn.)
var ConsulateDreadnought = newConsulateDreadnought

func newConsulateDreadnought() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Consulate Dreadnought",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 11}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(6),
			},
			OracleText: `
			Crew 6 (Tap any number of creatures you control with total power 6 or more: This Vehicle becomes an artifact creature until end of turn.)
		`,
		},
	}
}
