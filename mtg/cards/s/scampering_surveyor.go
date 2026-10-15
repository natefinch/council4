package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScamperingSurveyor is the card definition for Scampering Surveyor.
//
// Type: Artifact Creature — Gnome
// Cost: {4}
//
// Oracle text:
//
//	When this creature enters, search your library for a basic land card or Cave card, put it onto the battlefield tapped, then shuffle.
var ScamperingSurveyor = newScamperingSurveyor

func newScamperingSurveyor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scampering Surveyor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gnome},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, search your library for a basic land card or Cave card, put it onto the battlefield tapped, then shuffle.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:   zone.Library,
												Destination:  zone.Battlefield,
												Filter:       game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}}, game.Selection{SubtypesAny: []types.Sub{types.Cave}}}},
												EntersTapped: true,
											},
											Amount: game.Fixed(1),
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
			When this creature enters, search your library for a basic land card or Cave card, put it onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
