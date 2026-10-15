package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SunQuanLordOfWu is the card definition for Sun Quan, Lord of Wu.
//
// Type: Legendary Creature — Human Soldier
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Creatures you control have horsemanship. (They can't be blocked except by creatures with horsemanship.)
var SunQuanLordOfWu = newSunQuanLordOfWu

func newSunQuanLordOfWu() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sun Quan, Lord of Wu",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have horsemanship. (They can't be blocked except by creatures with horsemanship.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Horsemanship},
						},
					},
				},
			},
			OracleText: `
			Creatures you control have horsemanship. (They can't be blocked except by creatures with horsemanship.)
		`,
		},
	}
}
