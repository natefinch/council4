package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SoulShepherd is the card definition for Soul Shepherd.
//
// Type: Creature — Human Cleric
// Cost: {1}{W}
//
// Oracle text:
//
//	{W}, Exile a creature card from your graveyard: You gain 1 life.
var SoulShepherd = newSoulShepherd

func newSoulShepherd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Soul Shepherd",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}, Exile a creature card from your graveyard: You gain 1 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:          cost.AdditionalExile,
							Text:          "Exile a creature card from your graveyard",
							Amount:        1,
							MatchCardType: true,
							CardType:      types.Creature,
							Source:        zone.Graveyard,
							CounterKind:   counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
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
			OracleText: `
			{W}, Exile a creature card from your graveyard: You gain 1 life.
		`,
		},
	}
}
