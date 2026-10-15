package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VividFlyingFish is the card definition for Vivid Flying Fish.
//
// Type: Creature — Fish Lizard
// Cost: {1}{U}
//
// Oracle text:
//
//	This creature has flying as long as it's attacking. (It can't be blocked except by creatures with flying or reach.)
var VividFlyingFish = newVividFlyingFish

func newVividFlyingFish() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vivid Flying Fish",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fish, types.Lizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has flying as long as it's attacking. (It can't be blocked except by creatures with flying or reach.)",
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
							AddKeywords:    []game.Keyword{game.Flying},
						},
					},
				},
			},
			OracleText: `
			This creature has flying as long as it's attacking. (It can't be blocked except by creatures with flying or reach.)
		`,
		},
	}
}
