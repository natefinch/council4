package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PaladinOfPredation is the card definition for Paladin of Predation.
//
// Type: Creature — Phyrexian Knight
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Toxic 6 (Players dealt combat damage by this creature also get six poison counters.)
//	This creature can't be blocked by creatures with power 2 or less.
var PaladinOfPredation = newPaladinOfPredation

func newPaladinOfPredation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Paladin of Predation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Knight},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Toxic 6",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 6,
						},
					},
				},
				game.StaticAbility{
					Text: "This creature can't be blocked by creatures with power 2 or less.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			OracleText: `
			Toxic 6 (Players dealt combat damage by this creature also get six poison counters.)
			This creature can't be blocked by creatures with power 2 or less.
		`,
		},
	}
}
