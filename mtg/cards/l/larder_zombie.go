package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LarderZombie is the card definition for Larder Zombie.
//
// Type: Creature — Zombie
// Cost: {U}
//
// Oracle text:
//
//	Defender
//	Tap three untapped creatures you control: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
var LarderZombie = newLarderZombie

func newLarderZombie() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Larder Zombie",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap three untapped creatures you control: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "Tap three untapped creatures you control",
							Amount:             3,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Surveil{
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
			Defender
			Tap three untapped creatures you control: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
		`,
		},
	}
}
