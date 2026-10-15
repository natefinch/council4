package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LordOfTheUnreal is the card definition for Lord of the Unreal.
//
// Type: Creature — Human Wizard
// Cost: {U}{U}
//
// Oracle text:
//
//	Illusion creatures you control get +1/+1 and have hexproof. (They can't be the targets of spells or abilities your opponents control.)
var LordOfTheUnreal = newLordOfTheUnreal

func newLordOfTheUnreal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Lord of the Unreal",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Illusion creatures you control get +1/+1 and have hexproof. (They can't be the targets of spells or abilities your opponents control.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Illusion")}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Illusion")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Hexproof},
						},
					},
				},
			},
			OracleText: `
			Illusion creatures you control get +1/+1 and have hexproof. (They can't be the targets of spells or abilities your opponents control.)
		`,
		},
	}
}
