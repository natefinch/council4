package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Immerwolf is the card definition for Immerwolf.
//
// Type: Creature — Wolf
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
//	Each other creature you control that's a Wolf or a Werewolf gets +1/+1.
//	Non-Human Werewolves you control can't transform.
var Immerwolf = newImmerwolf

func newImmerwolf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Immerwolf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.IntimidateStaticBody,
				game.StaticAbility{
					Text: "Each other creature you control that's a Wolf or a Werewolf gets +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Wolf"), types.Sub("Werewolf")}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
				game.StaticAbility{
					Text: "Non-Human Werewolves you control can't transform.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantTransform,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							AffectedSelection:  game.Selection{SubtypesAny: []types.Sub{types.Werewolf}, ExcludedSubtype: types.Human},
						},
					},
				},
			},
			OracleText: `
			Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
			Each other creature you control that's a Wolf or a Werewolf gets +1/+1.
			Non-Human Werewolves you control can't transform.
		`,
		},
	}
}
