package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThornedMoloch is the card definition for Thorned Moloch.
//
// Type: Creature — Lizard
// Cost: {2}{R}
//
// Oracle text:
//
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	This creature has first strike as long as it's attacking.
var ThornedMoloch = newThornedMoloch

func newThornedMoloch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Thorned Moloch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
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
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			This creature has first strike as long as it's attacking.
		`,
		},
	}
}
