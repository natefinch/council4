package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GlitchGhostSurveyor is the card definition for Glitch Ghost Surveyor.
//
// Type: Creature — Spirit Scout
// Cost: {2}{U}
//
// Oracle text:
//
//	Flying
//	Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
//	Max speed — {3}, Exile this card from your graveyard: Draw a card.
var GlitchGhostSurveyor = newGlitchGhostSurveyor

func newGlitchGhostSurveyor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glitch Ghost Surveyor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Scout},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Max speed — {3}, Exile this card from your graveyard: Draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile this card from your graveyard",
							Amount:      1,
							Source:      zone.Graveyard,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					ActivationCondition: opt.Val(game.Condition{
						SourceCounterKind:     counter.PlusOnePlusOne,
						ControllerHasMaxSpeed: true,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.StartEnginesTriggeredBody,
			},
			OracleText: `
			Flying
			Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
			Max speed — {3}, Exile this card from your graveyard: Draw a card.
		`,
		},
	}
}
