package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrderOfSacredDusk is the card definition for Order of Sacred Dusk.
//
// Type: Creature — Vampire Knight
// Cost: {6}{W}{B}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	Flying, lifelink, haste
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
//	Other Vampires you control have exalted.
var OrderOfSacredDusk = newOrderOfSacredDusk

func newOrderOfSacredDusk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Order of Sacred Dusk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Knight},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
				game.HasteStaticBody,
				game.ExaltedStaticBody,
				game.StaticAbility{
					Text: "Other Vampires you control have exalted.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Vampire")}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Exalted},
						},
					},
				},
			},
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			Flying, lifelink, haste
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
			Other Vampires you control have exalted.
		`,
		},
	}
}
