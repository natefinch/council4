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

// ColossalGraveReaver is the card definition for Colossal Grave-Reaver.
//
// Type: Creature — Dragon
// Cost: {6}{B}{G}
//
// Oracle text:
//
//	Flying
//	Whenever this creature enters or attacks, mill three cards.
//	Whenever one or more creature cards are put into your graveyard from your library, put one of them onto the battlefield.
var ColossalGraveReaver = newColossalGraveReaver

func newColossalGraveReaver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Colossal Grave-Reaver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature enters or attacks, mill three cards.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							UnionEvent:      game.EventAttackerDeclared,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(3),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Whenever one or more creature cards are put into your graveyard from your library, put one of them onto the battlefield.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MatchFromZone:    true,
							FromZone:         zone.Library,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							StackObjectKind:  game.StackSpell,
							OneOrMore:        true,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MassReturnFromGraveyard{
											Player:           game.ControllerReference(),
											Destination:      zone.Battlefield,
											FromTriggerBatch: true,
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
			Flying
			Whenever this creature enters or attacks, mill three cards.
			Whenever one or more creature cards are put into your graveyard from your library, put one of them onto the battlefield.
		`,
		},
	}
}
