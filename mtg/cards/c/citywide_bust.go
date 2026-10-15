package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CitywideBust is the card definition for Citywide Bust.
//
// Type: Sorcery
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Destroy all creatures with toughness 4 or greater.
var CitywideBust = newCitywideBust

func newCitywideBust() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Citywide Bust",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Toughness: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures with toughness 4 or greater.
		`,
		},
	}
}
