package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VorinclexVoiceOfHunger is the card definition for Vorinclex, Voice of Hunger.
//
// Type: Legendary Creature — Phyrexian Praetor
// Cost: {6}{G}{G}
//
// Oracle text:
//
//	Trample
//	Whenever you tap a land for mana, add one mana of any type that land produced.
//	Whenever an opponent taps a land for mana, that land doesn't untap during its controller's next untap step.
var VorinclexVoiceOfHunger = newVorinclexVoiceOfHunger()

func newVorinclexVoiceOfHunger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vorinclex, Voice of Hunger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Praetor},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentTapped,
							Controller:           game.TriggerControllerYou,
							RequireTappedForMana: true,
							SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:        game.ResolutionChoiceMana,
										Prompt:      "Choose a type of mana that land produced",
										ColorSource: game.ResolutionChoiceColorSourceTriggerLandProduced,
									},
									PublishChoice: game.ChoiceKey("oracle-trigger-land-produced-type"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-trigger-land-produced-type"),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentTapped,
							Controller:           game.TriggerControllerOpponent,
							RequireTappedForMana: true,
							SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SkipNextUntap{
									Object: game.EventPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever you tap a land for mana, add one mana of any type that land produced.
			Whenever an opponent taps a land for mana, that land doesn't untap during its controller's next untap step.
		`,
		},
	}
}
