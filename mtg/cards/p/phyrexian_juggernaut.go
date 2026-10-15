package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianJuggernaut is the card definition for Phyrexian Juggernaut.
//
// Type: Artifact Creature — Phyrexian Juggernaut
// Cost: {6}
//
// Oracle text:
//
//	Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	This creature attacks each combat if able.
var PhyrexianJuggernaut = newPhyrexianJuggernaut

func newPhyrexianJuggernaut() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Phyrexian Juggernaut",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Juggernaut},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.InfectStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
			This creature attacks each combat if able.
		`,
		},
	}
}
