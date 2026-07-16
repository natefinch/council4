package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GiltLeafArchdruid is the card definition for Gilt-Leaf Archdruid.
//
// Type: Creature — Elf Druid
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Whenever you cast a Druid spell, you may draw a card.
//	Tap seven untapped Druids you control: Gain control of all lands target player controls.
var GiltLeafArchdruid = newGiltLeafArchdruid

func newGiltLeafArchdruid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gilt-Leaf Archdruid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap seven untapped Druids you control: Gain control of all lands target player controls.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap seven untapped Druids you control",
							Amount:      7,
							SubtypesAny: cost.SubtypeSet{types.Druid},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:         game.LayerControl,
											NewController: opt.Val(game.Player1),
											Group:         game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Land}}),
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Druid")}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast a Druid spell, you may draw a card.
			Tap seven untapped Druids you control: Gain control of all lands target player controls.
		`,
		},
	}
}
