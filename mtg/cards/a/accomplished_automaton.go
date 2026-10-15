package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AccomplishedAutomaton is the card definition for Accomplished Automaton.
//
// Type: Artifact Creature — Construct
// Cost: {7}
//
// Oracle text:
//
//	Fabricate 1 (When this creature enters, put a +1/+1 counter on it or create a 1/1 colorless Servo artifact creature token.)
var AccomplishedAutomaton = newAccomplishedAutomaton

func newAccomplishedAutomaton() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Accomplished Automaton",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.FabricateTriggeredAbility(1),
			},
			OracleText: `
			Fabricate 1 (When this creature enters, put a +1/+1 counter on it or create a 1/1 colorless Servo artifact creature token.)
		`,
		},
	}
}
