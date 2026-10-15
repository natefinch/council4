package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CacklingSlasher is the card definition for Cackling Slasher.
//
// Type: Creature — Human Assassin
// Cost: {3}{B}
//
// Oracle text:
//
//	Deathtouch
//	This creature enters with a +1/+1 counter on it if a creature died this turn.
var CacklingSlasher = newCacklingSlasher

func newCacklingSlasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cackling Slasher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Assassin},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("This creature enters with a +1/+1 counter on it if a creature died this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:            game.EventPermanentDied,
						SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						StackObjectKind:  game.StackSpell,
						CounterKind:      counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Deathtouch
			This creature enters with a +1/+1 counter on it if a creature died this turn.
		`,
		},
	}
}
