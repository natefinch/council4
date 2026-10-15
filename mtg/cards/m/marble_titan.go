package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MarbleTitan is the card definition for Marble Titan.
//
// Type: Creature — Giant
// Cost: {3}{W}
//
// Oracle text:
//
//	Creatures with power 3 or greater don't untap during their controllers' untap steps.
var MarbleTitan = newMarbleTitan

func newMarbleTitan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Marble Titan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures with power 3 or greater don't untap during their controllers' untap steps.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
							AffectedSelection: game.Selection{Power: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 3,
							})},
						},
					},
				},
			},
			OracleText: `
			Creatures with power 3 or greater don't untap during their controllers' untap steps.
		`,
		},
	}
}
