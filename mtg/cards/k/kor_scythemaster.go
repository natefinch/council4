package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KorScythemaster is the card definition for Kor Scythemaster.
//
// Type: Creature — Kor Soldier Ally
// Cost: {2}{W}
//
// Oracle text:
//
//	This creature has first strike as long as it's attacking.
var KorScythemaster = newKorScythemaster

func newKorScythemaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kor Scythemaster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Soldier, types.Ally},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has first strike as long as it's attacking.",
					Condition: opt.Val(game.Condition{
						Text:              "as long as it's attacking",
						Object:            opt.Val(game.SourcePermanentReference()),
						ObjectMatches:     opt.Val(game.Selection{CombatState: game.CombatStateAttacking}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			This creature has first strike as long as it's attacking.
		`,
		},
	}
}
