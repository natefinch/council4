package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HorribleHordes is the card definition for Horrible Hordes.
//
// Type: Artifact Creature — Spirit
// Cost: {3}
//
// Oracle text:
//
//	Rampage 1 (Whenever this creature becomes blocked, it gets +1/+1 until end of turn for each creature blocking it beyond the first.)
var HorribleHordes = newHorribleHordes

func newHorribleHordes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Horrible Hordes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.RampageTriggeredAbility(1),
			},
			OracleText: `
			Rampage 1 (Whenever this creature becomes blocked, it gets +1/+1 until end of turn for each creature blocking it beyond the first.)
		`,
		},
	}
}
