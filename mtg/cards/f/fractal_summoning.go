package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FractalSummoning is the card definition for Fractal Summoning.
//
// Type: Sorcery — Lesson
// Cost: {X}{G/U}{G/U}
//
// Oracle text:
//
//	Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it.
var FractalSummoning = newFractalSummoning()

func newFractalSummoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Fractal Summoning",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:   []color.Color{color.Green, color.Blue},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(1),
							Source:        game.TokenDef(fractalSummoningToken),
							PublishLinked: game.LinkedKey("created-token"),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Object:      game.LinkedObjectReference("created-token"),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it.
		`,
		},
	}
}

var fractalSummoningToken = newFractalSummoningToken()

func newFractalSummoningToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Fractal",
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fractal},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		},
	}
}
