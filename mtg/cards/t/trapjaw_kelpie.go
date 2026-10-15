package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrapjawKelpie is the card definition for Trapjaw Kelpie.
//
// Type: Creature — Beast
// Cost: {4}{G/U}{G/U}
//
// Oracle text:
//
//	Flash
//	Persist (When this creature dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)
var TrapjawKelpie = newTrapjawKelpie

func newTrapjawKelpie() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Trapjaw Kelpie",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.PersistTriggeredBody,
			},
			OracleText: `
			Flash
			Persist (When this creature dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)
		`,
		},
	}
}
