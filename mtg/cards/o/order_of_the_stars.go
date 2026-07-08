package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrderOfTheStars is the card definition for Order of the Stars.
//
// Type: Creature — Human Cleric
// Cost: {W}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	As this creature enters, choose a color.
//	This creature has protection from the chosen color.
var OrderOfTheStars = newOrderOfTheStars

func newOrderOfTheStars() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Order of the Stars",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddAbilities: []game.Ability{
								new(game.ProtectionFromChosenColorStaticAbility()),
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this creature enters, choose a color."),
			},
			OracleText: `
			Defender (This creature can't attack.)
			As this creature enters, choose a color.
			This creature has protection from the chosen color.
		`,
		},
	}
}
