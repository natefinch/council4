package x

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// XerexStrobeKnight is the card definition for Xerex Strobe-Knight.
//
// Type: Creature — Human Knight
// Cost: {2}{U}
//
// Oracle text:
//
//	Flying, vigilance
//	{T}: Create a 2/2 white and blue Knight creature token with vigilance. Activate only if you've cast two or more spells this turn.
var XerexStrobeKnight = newXerexStrobeKnight()

func newXerexStrobeKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Xerex Strobe-Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Create a 2/2 white and blue Knight creature token with vigilance. Activate only if you've cast two or more spells this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:      game.EventSpellCast,
							Controller: game.TriggerControllerYou,
						}, Window: game.EventHistoryCurrentTurn, MinCount: 2}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(xerexStrobeKnightToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, vigilance
			{T}: Create a 2/2 white and blue Knight creature token with vigilance. Activate only if you've cast two or more spells this turn.
		`,
		},
	}
}

var xerexStrobeKnightToken = newXerexStrobeKnightToken()

func newXerexStrobeKnightToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.White, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}
