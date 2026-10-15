package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KozilekSTranslator is the card definition for Kozilek's Translator.
//
// Type: Creature — Eldrazi Drone
// Cost: {4}{B}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Pay 1 life: Add {C}. Activate only once each turn. ({C} represents colorless mana.)
var KozilekSTranslator = newKozilekSTranslator

func newKozilekSTranslator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Kozilek's Translator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Pay 1 life: Add {C}. Activate only once each turn. ({C} represents colorless mana.)",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalPayLife,
							Text:        "Pay 1 life",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Timing: game.OncePerTurn,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
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
			OracleText: `
			Devoid (This card has no color.)
			Pay 1 life: Add {C}. Activate only once each turn. ({C} represents colorless mana.)
		`,
		},
	}
}
