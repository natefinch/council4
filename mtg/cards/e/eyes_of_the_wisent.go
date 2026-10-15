package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EyesOfTheWisent is the card definition for Eyes of the Wisent.
//
// Type: Kindred Enchantment — Elemental
// Cost: {1}{G}
//
// Oracle text:
//
//	Whenever an opponent casts a blue spell during your turn, you may create a 4/4 green Elemental creature token.
var EyesOfTheWisent = newEyesOfTheWisent

func newEyesOfTheWisent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Eyes of the Wisent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Kindred, types.Enchantment},
			Subtypes: []types.Sub{types.Elemental},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent casts a blue spell during your turn, you may create a 4/4 green Elemental creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerOpponent,
							CardSelection:   game.Selection{ColorsAny: []color.Color{color.Blue}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							CastDuringTurn:  game.TriggerTurnYours,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(eyesOfTheWisentToken),
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
			Whenever an opponent casts a blue spell during your turn, you may create a 4/4 green Elemental creature token.
		`,
		},
	}
}

var eyesOfTheWisentToken = newEyesOfTheWisentToken()

func newEyesOfTheWisentToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
