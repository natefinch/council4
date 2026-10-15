package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JolraelSCentaur is the card definition for Jolrael's Centaur.
//
// Type: Creature — Centaur Archer
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Shroud (This creature can't be the target of spells or abilities.)
//	Flanking (Whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)
var JolraelSCentaur = newJolraelSCentaur

func newJolraelSCentaur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Jolrael's Centaur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Centaur, types.Archer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ShroudStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.FlankingTriggeredBody,
			},
			OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
			Flanking (Whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)
		`,
		},
	}
}
