package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StirTheSands is the card definition for Stir the Sands.
//
// Type: Sorcery
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Create three 2/2 black Zombie creature tokens.
//	Cycling {3}{B} ({3}{B}, Discard this card: Draw a card.)
//	When you cycle this card, create a 2/2 black Zombie creature token.
var StirTheSands = newStirTheSands

func newStirTheSands() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Stir the Sands",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3), cost.B}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you cycle this card, create a 2/2 black Zombie creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventCycled,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(stirTheSandsToken),
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
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(stirTheSandsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 2/2 black Zombie creature tokens.
			Cycling {3}{B} ({3}{B}, Discard this card: Draw a card.)
			When you cycle this card, create a 2/2 black Zombie creature token.
		`,
		},
	}
}

var stirTheSandsToken = newStirTheSandsToken()

func newStirTheSandsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
