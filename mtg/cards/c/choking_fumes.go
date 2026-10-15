package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChokingFumes is the card definition for Choking Fumes.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Put a -1/-1 counter on each attacking creature.
var ChokingFumes = newChokingFumes

func newChokingFumes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Choking Fumes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							CounterKind: counter.MinusOneMinusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put a -1/-1 counter on each attacking creature.
		`,
		},
	}
}
