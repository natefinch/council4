package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WingedHiveTyrant is the card definition for Winged Hive Tyrant.
//
// Type: Creature — Tyranid
// Cost: {3}{U}{R}
//
// Oracle text:
//
//	Flying, haste
//	The Will of the Hive Mind — Other creatures you control with counters on them have flying and haste.
var WingedHiveTyrant = newWingedHiveTyrant

func newWingedHiveTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Winged Hive Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
				game.StaticAbility{
					Text: "The Will of the Hive Mind — Other creatures you control with counters on them have flying and haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchAnyCounter: true}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying, game.Haste},
						},
					},
				},
			},
			OracleText: `
			Flying, haste
			The Will of the Hive Mind — Other creatures you control with counters on them have flying and haste.
		`,
		},
	}
}
