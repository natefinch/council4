package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReveilleSquad is the card definition for Reveille Squad.
//
// Type: Creature — Human Rebel
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Whenever one or more creatures attack you, if this creature is untapped, you may untap all creatures you control.
var ReveilleSquad = newReveilleSquad

func newReveilleSquad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Reveille Squad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rebel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever one or more creatures attack you, if this creature is untapped, you may untap all creatures you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							AttackRecipient:  game.AttackRecipientPlayer,
							OneOrMore:        true,
							CounterKind:      counter.PlusOnePlusOne,
						},
						InterveningIf: "if this creature is untapped",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if this creature is untapped",
							Object:            opt.Val(game.SourcePermanentReference()),
							ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriFalse}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
										},
										Optional: true,
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
			Whenever one or more creatures attack you, if this creature is untapped, you may untap all creatures you control.
		`,
		},
	}
}
