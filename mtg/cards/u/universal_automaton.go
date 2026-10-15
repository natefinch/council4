package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UniversalAutomaton is the card definition for Universal Automaton.
//
// Type: Artifact Creature — Shapeshifter
// Cost: {1}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
var UniversalAutomaton = newUniversalAutomaton

func newUniversalAutomaton() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Universal Automaton",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
		`,
		},
	}
}
