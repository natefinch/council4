package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GorgerWurm is the card definition for Gorger Wurm.
//
// Type: Creature — Wurm
// Cost: {3}{R}{G}
//
// Oracle text:
//
//	Devour 1 (As this creature enters, you may sacrifice any number of creatures. It enters with that many +1/+1 counters on it.)
var GorgerWurm = newGorgerWurm

func newGorgerWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Gorger Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.DevourReplacement("As this creature enters, you may sacrifice any number of creatures, then it enters with 1 +1/+1 counters on it for each creature sacrificed.", 1),
			},
			OracleText: `
			Devour 1 (As this creature enters, you may sacrifice any number of creatures. It enters with that many +1/+1 counters on it.)
		`,
		},
	}
}
