package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThranTemporalGateway is the card definition for Thran Temporal Gateway.
//
// Type: Legendary Artifact
// Cost: {4}
//
// Oracle text:
//
//	{4}, {T}: You may put a historic permanent card from your hand onto the battlefield. (Artifacts, legendaries, and Sagas are historic.)
var ThranTemporalGateway = newThranTemporalGateway

func newThranTemporalGateway() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thran Temporal Gateway",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}, {T}: You may put a historic permanent card from your hand onto the battlefield. (Artifacts, legendaries, and Sagas are historic.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Hand,
											Filter:     game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Artifact}}, game.Selection{Supertypes: []types.Super{types.Legendary}}, game.Selection{SubtypesAny: []types.Sub{types.Saga}}}, RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Battlefield,
											},
											Prompt: "Choose a card to put onto the battlefield",
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
			{4}, {T}: You may put a historic permanent card from your hand onto the battlefield. (Artifacts, legendaries, and Sagas are historic.)
		`,
		},
	}
}
