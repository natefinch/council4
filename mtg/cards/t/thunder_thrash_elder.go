package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThunderThrashElder is the card definition for Thunder-Thrash Elder.
//
// Type: Creature — Lizard Warrior
// Cost: {2}{R}
//
// Oracle text:
//
//	Devour 3 (As this creature enters, you may sacrifice any number of creatures. It enters with three times that many +1/+1 counters on it.)
var ThunderThrashElder = newThunderThrashElder

func newThunderThrashElder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Thunder-Thrash Elder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.DevourReplacement("As this creature enters, you may sacrifice any number of creatures, then it enters with 3 +1/+1 counters on it for each creature sacrificed.", 3),
			},
			OracleText: `
			Devour 3 (As this creature enters, you may sacrifice any number of creatures. It enters with three times that many +1/+1 counters on it.)
		`,
		},
	}
}
