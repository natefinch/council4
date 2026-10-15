package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CallTheBloodline is the card definition for Call the Bloodline.
//
// Type: Enchantment
// Cost: {1}{B}
//
// Oracle text:
//
//	{1}, Discard a card: Create a 1/1 black Vampire Knight creature token with lifelink. Activate only once each turn.
var CallTheBloodline = newCallTheBloodline

func newCallTheBloodline() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Call the Bloodline",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, Discard a card: Create a 1/1 black Vampire Knight creature token with lifelink. Activate only once each turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard a card",
							Amount:      1,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(callTheBloodlineToken),
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
			{1}, Discard a card: Create a 1/1 black Vampire Knight creature token with lifelink. Activate only once each turn.
		`,
		},
	}
}

var callTheBloodlineToken = newCallTheBloodlineToken()

func newCallTheBloodlineToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Vampire Knight",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Knight},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
		},
	}
}
