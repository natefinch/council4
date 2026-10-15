package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GluttonousSlime is the card definition for Gluttonous Slime.
//
// Type: Creature — Ooze
// Cost: {2}{G}
//
// Oracle text:
//
//	Flash
//	Devour 1 (As this creature enters, you may sacrifice any number of creatures. It enters with that many +1/+1 counters on it.)
var GluttonousSlime = newGluttonousSlime

func newGluttonousSlime() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gluttonous Slime",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ooze},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DevourReplacement("As this creature enters, you may sacrifice any number of creatures, then it enters with 1 +1/+1 counters on it for each creature sacrificed.", 1),
			},
			OracleText: `
			Flash
			Devour 1 (As this creature enters, you may sacrifice any number of creatures. It enters with that many +1/+1 counters on it.)
		`,
		},
	}
}
