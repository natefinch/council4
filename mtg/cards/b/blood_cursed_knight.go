package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodCursedKnight is the card definition for Blood-Cursed Knight.
//
// Type: Creature — Vampire Knight
// Cost: {1}{W}{B}
//
// Oracle text:
//
//	As long as you control an enchantment, this creature gets +1/+1 and has lifelink. (Damage dealt by this creature also causes you to gain that much life.)
var BloodCursedKnight = newBloodCursedKnight

func newBloodCursedKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Blood-Cursed Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control an enchantment, this creature gets +1/+1 and has lifelink. (Damage dealt by this creature also causes you to gain that much life.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control an enchantment",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Lifelink},
						},
					},
				},
			},
			OracleText: `
			As long as you control an enchantment, this creature gets +1/+1 and has lifelink. (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
