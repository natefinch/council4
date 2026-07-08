package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoonlitLamenter is the card definition for Moonlit Lamenter.
//
// Type: Creature — Treefolk Cleric
// Cost: {2}{W}
//
// Oracle text:
//
//	This creature enters with a -1/-1 counter on it.
//	{1}{W}, Remove a counter from this creature: Draw a card. Activate only as a sorcery.
var MoonlitLamenter = newMoonlitLamenter

func newMoonlitLamenter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Moonlit Lamenter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{W}, Remove a counter from this creature: Draw a card. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalRemoveCounter,
							Text:           "Remove a counter from this creature",
							Amount:         1,
							AnyCounterKind: true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a -1/-1 counter on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 1}),
			},
			OracleText: `
			This creature enters with a -1/-1 counter on it.
			{1}{W}, Remove a counter from this creature: Draw a card. Activate only as a sorcery.
		`,
		},
	}
}
