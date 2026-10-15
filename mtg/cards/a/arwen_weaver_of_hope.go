package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArwenWeaverOfHope is the card definition for Arwen, Weaver of Hope.
//
// Type: Legendary Creature — Elf Noble
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Each other creature you control enters with a number of additional +1/+1 counters on it equal to Arwen's toughness.
var ArwenWeaverOfHope = newArwenWeaverOfHope

func newArwenWeaverOfHope() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Arwen, Weaver of Hope",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Noble},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersGroupReplacement("Each other creature you control enters with a number of additional +1/+1 counters on it equal to Arwen's toughness.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountObjectToughness,
					Multiplier: 1,
					Object:     game.SourcePermanentReference(),
				})}),
			},
			OracleText: `
			Each other creature you control enters with a number of additional +1/+1 counters on it equal to Arwen's toughness.
		`,
		},
	}
}
