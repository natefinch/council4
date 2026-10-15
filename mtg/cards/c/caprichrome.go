package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Caprichrome is the card definition for Caprichrome.
//
// Type: Artifact Creature — Goat
// Cost: {3}{W}
//
// Oracle text:
//
//	Flash
//	Vigilance
//	Devour artifact 1 (As this creature enters, you may sacrifice any number of artifacts. It enters with that many +1/+1 counters on it.)
var Caprichrome = newCaprichrome

func newCaprichrome() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Caprichrome",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Goat},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.VigilanceStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DevourTypeReplacement("As this creature enters, you may sacrifice any number of artifacts, then it enters with 1 +1/+1 counters on it for each artifact sacrificed.", 1, types.Artifact),
			},
			OracleText: `
			Flash
			Vigilance
			Devour artifact 1 (As this creature enters, you may sacrifice any number of artifacts. It enters with that many +1/+1 counters on it.)
		`,
		},
	}
}
