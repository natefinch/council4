package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Chromanticore is the card definition for Chromanticore.
//
// Type: Enchantment Creature — Manticore
// Cost: {W}{U}{B}{R}{G}
//
// Oracle text:
//
//	Bestow {2}{W}{U}{B}{R}{G} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	Flying, first strike, vigilance, trample, lifelink
//	Enchanted creature gets +4/+4 and has flying, first strike, vigilance, trample, and lifelink.
var Chromanticore = newChromanticore

func newChromanticore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Chromanticore",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Manticore},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.BestowStaticAbility(cost.Mana{cost.O(2), cost.W, cost.U, cost.B, cost.R, cost.G}, &game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     4,
							ToughnessDelta: 4,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Flying,
								game.FirstStrike,
								game.Vigilance,
								game.Trample,
								game.Lifelink,
							},
						},
					},
				},
			},
			OracleText: `
			Bestow {2}{W}{U}{B}{R}{G} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
			Flying, first strike, vigilance, trample, lifelink
			Enchanted creature gets +4/+4 and has flying, first strike, vigilance, trample, and lifelink.
		`,
		},
	}
}
