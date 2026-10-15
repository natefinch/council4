package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RetributionOfTheMeek is the card definition for Retribution of the Meek.
//
// Type: Sorcery
// Cost: {2}{W}
//
// Oracle text:
//
//	Destroy all creatures with power 4 or greater. They can't be regenerated.
var RetributionOfTheMeek = newRetributionOfTheMeek

func newRetributionOfTheMeek() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Retribution of the Meek",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group:               game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures with power 4 or greater. They can't be regenerated.
		`,
		},
	}
}
