package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CelestialCrusader is the card definition for Celestial Crusader.
//
// Type: Creature — Spirit
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flash
//	Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
//	Flying
//	Other white creatures get +1/+1.
var CelestialCrusader = newCelestialCrusader

func newCelestialCrusader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Celestial Crusader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.SplitSecondStaticBody,
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Other white creatures get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.White}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Flash
			Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
			Flying
			Other white creatures get +1/+1.
		`,
		},
	}
}
