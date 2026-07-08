package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GarenbrigPaladin is the card definition for Garenbrig Paladin.
//
// Type: Creature — Giant Knight
// Cost: {4}{G}
//
// Oracle text:
//
//	Adamant — If at least three green mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
//	This creature can't be blocked by creatures with power 2 or less.
var GarenbrigPaladin = newGarenbrigPaladin

func newGarenbrigPaladin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Garenbrig Paladin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Knight},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Adamant — If at least three green mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.", &game.Condition{
					SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Green, Count: 3},
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Adamant — If at least three green mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
			This creature can't be blocked by creatures with power 2 or less.
		`,
		},
	}
}
