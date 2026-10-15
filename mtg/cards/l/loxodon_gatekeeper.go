package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LoxodonGatekeeper is the card definition for Loxodon Gatekeeper.
//
// Type: Creature — Elephant Soldier
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Artifacts, creatures, and lands your opponents control enter tapped.
var LoxodonGatekeeper = newLoxodonGatekeeper

func newLoxodonGatekeeper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Loxodon Gatekeeper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedGroupReplacement("Artifacts, creatures, and lands your opponents control enter tapped.", game.TriggerControllerOpponent, types.Artifact, types.Creature, types.Land),
			},
			OracleText: `
			Artifacts, creatures, and lands your opponents control enter tapped.
		`,
		},
	}
}
