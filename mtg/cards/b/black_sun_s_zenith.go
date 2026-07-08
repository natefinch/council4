package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlackSunSZenith is the card definition for Black Sun's Zenith.
//
// Type: Sorcery
// Cost: {X}{B}{B}
//
// Oracle text:
//
//	Put X -1/-1 counters on each creature. Shuffle Black Sun's Zenith into its owner's library.
var BlackSunSZenith = newBlackSunSZenith

func newBlackSunSZenith() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Black Sun's Zenith",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							CounterKind: counter.MinusOneMinusOne,
						},
					},
					{
						Primitive: game.ShuffleSpellIntoLibrary{},
					},
				},
			}.Ability()),
			OracleText: `
			Put X -1/-1 counters on each creature. Shuffle Black Sun's Zenith into its owner's library.
		`,
		},
	}
}
