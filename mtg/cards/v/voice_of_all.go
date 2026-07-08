package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VoiceOfAll is the card definition for Voice of All.
//
// Type: Creature — Angel
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flying
//	As this creature enters, choose a color.
//	This creature has protection from the chosen color.
var VoiceOfAll = newVoiceOfAll

func newVoiceOfAll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Voice of All",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
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
			Flying
			As this creature enters, choose a color.
			This creature has protection from the chosen color.
		`,
		},
	}
}
