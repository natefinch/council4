package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CullingRitual is the card definition for Culling Ritual.
//
// Type: Sorcery
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	Destroy each nonland permanent with mana value 2 or less. Add {B} or {G} for each permanent destroyed this way.
var CullingRitual = newCullingRitual

func newCullingRitual() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Culling Ritual",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
						},
						PublishResult: game.ResultKey("destroyed-this-way"),
					},
					{
						Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("destroyed-this-way"),
							}),
							CombinationColors: []mana.Color{mana.B, mana.G},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy each nonland permanent with mana value 2 or less. Add {B} or {G} for each permanent destroyed this way.
		`,
		},
	}
}
