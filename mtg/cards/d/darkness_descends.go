package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarknessDescends is the card definition for Darkness Descends.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Put two -1/-1 counters on each creature.
var DarknessDescends = newDarknessDescends

func newDarknessDescends() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Darkness Descends",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							CounterKind: counter.MinusOneMinusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put two -1/-1 counters on each creature.
		`,
		},
	}
}
