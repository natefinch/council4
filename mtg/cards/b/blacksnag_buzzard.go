package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlacksnagBuzzard is the card definition for Blacksnag Buzzard.
//
// Type: Creature — Bird
// Cost: {2}{B}
//
// Oracle text:
//
//	Flying
//	This creature enters with a +1/+1 counter on it if a creature died this turn.
//	Plot {1}{B} (You may pay {1}{B} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
var BlacksnagBuzzard = newBlacksnagBuzzard

func newBlacksnagBuzzard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blacksnag Buzzard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Plot {1}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.PlotKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.B,
							},
						},
					},
				},
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
			Flying
			This creature enters with a +1/+1 counter on it if a creature died this turn.
			Plot {1}{B} (You may pay {1}{B} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
		`,
		},
	}
}
