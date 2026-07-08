package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DwarfholdChampion is the card definition for Dwarfhold Champion.
//
// Type: Creature — Dwarf Warrior
// Cost: {1}{W}
//
// Oracle text:
//
//	As long as this creature is equipped, it gets +0/+2.
var DwarfholdChampion = newDwarfholdChampion

func newDwarfholdChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Dwarfhold Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dwarf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchEquipped: true}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			As long as this creature is equipped, it gets +0/+2.
		`,
		},
	}
}
