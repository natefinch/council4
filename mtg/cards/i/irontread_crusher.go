package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IrontreadCrusher is the card definition for Irontread Crusher.
//
// Type: Artifact — Vehicle
// Cost: {4}
//
// Oracle text:
//
//	Crew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)
var IrontreadCrusher = newIrontreadCrusher

func newIrontreadCrusher() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Irontread Crusher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(3),
			},
			OracleText: `
			Crew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)
		`,
		},
	}
}
