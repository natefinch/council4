package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LyraDawnbringer is the card definition for Lyra Dawnbringer.
//
// Type: Legendary Creature — Angel
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying
//	First strike (This creature deals combat damage before creatures without first strike.)
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	Other Angels you control get +1/+1 and have lifelink.
var LyraDawnbringer = newLyraDawnbringer

func newLyraDawnbringer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Lyra Dawnbringer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Other Angels you control get +1/+1 and have lifelink.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Angel")}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Angel")}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Lifelink},
						},
					},
				},
			},
			OracleText: `
			Flying
			First strike (This creature deals combat damage before creatures without first strike.)
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			Other Angels you control get +1/+1 and have lifelink.
		`,
		},
	}
}
