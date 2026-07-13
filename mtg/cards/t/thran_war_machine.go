package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThranWarMachine is the card definition for Thran War Machine.
//
// Type: Artifact Creature — Construct
// Cost: {4}
//
// Oracle text:
//
//	Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	This creature attacks each combat if able.
var ThranWarMachine = newThranWarMachine

func newThranWarMachine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thran War Machine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(4)}),
			},
			OracleText: `
			Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			This creature attacks each combat if able.
		`,
		},
	}
}
