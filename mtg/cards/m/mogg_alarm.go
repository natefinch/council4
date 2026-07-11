package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoggAlarm is the card definition for Mogg Alarm.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	You may sacrifice two Mountains rather than pay this spell's mana cost.
//	Create two 1/1 red Goblin creature tokens.
var MoggAlarm = newMoggAlarm

func newMoggAlarm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mogg Alarm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice two Mountains",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "sacrifice two Mountains",
							Amount:      2,
							SubtypesAny: cost.SubtypeSet{types.Mountain},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(moggAlarmToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may sacrifice two Mountains rather than pay this spell's mana cost.
			Create two 1/1 red Goblin creature tokens.
		`,
		},
	}
}

var moggAlarmToken = newMoggAlarmToken()

func newMoggAlarmToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
