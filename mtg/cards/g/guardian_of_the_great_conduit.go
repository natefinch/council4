package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuardianOfTheGreatConduit is the card definition for Guardian of the Great Conduit.
//
// Type: Creature — Elemental
// Cost: {3}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	As long as you control a Nissa planeswalker, this creature gets +2/+0 and has vigilance. (Attacking doesn't cause it to tap.)
var GuardianOfTheGreatConduit = newGuardianOfTheGreatConduit

func newGuardianOfTheGreatConduit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Guardian of the Great Conduit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.StaticAbility{
					Text: "As long as you control a Nissa planeswalker, this creature gets +2/+0 and has vigilance. (Attacking doesn't cause it to tap.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Nissa planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Nissa}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Vigilance},
						},
					},
				},
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			As long as you control a Nissa planeswalker, this creature gets +2/+0 and has vigilance. (Attacking doesn't cause it to tap.)
		`,
		},
	}
}
