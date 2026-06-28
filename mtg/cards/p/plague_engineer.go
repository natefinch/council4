package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PlagueEngineer is the card definition for Plague Engineer.
//
// Type: Creature — Phyrexian Carrier
// Cost: {2}{B}
//
// Oracle text:
//
//	Deathtouch
//	As this creature enters, choose a creature type.
//	Creatures of the chosen type your opponents control get -1/-1.
var PlagueEngineer = newPlagueEngineer()

func newPlagueEngineer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Plague Engineer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Carrier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry, Controller: game.ControllerOpponent}),
							PowerDelta:     -1,
							ToughnessDelta: -1,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this creature enters, choose a creature type."),
			},
			OracleText: `
			Deathtouch
			As this creature enters, choose a creature type.
			Creatures of the chosen type your opponents control get -1/-1.
		`,
		},
	}
}
