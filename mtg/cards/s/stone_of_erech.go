package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StoneOfErech is the card definition for Stone of Erech.
//
// Type: Legendary Artifact
// Cost: {1}
//
// Oracle text:
//
//	If a creature an opponent controls would die, exile it instead.
//	{2}, {T}, Sacrifice Stone of Erech: Exile target player's graveyard. Draw a card.
var StoneOfErech = newStoneOfErech

func newStoneOfErech() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Stone of Erech",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Sacrifice Stone of Erech: Exile target player's graveyard. Draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice Stone of Erech",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target player's graveyard",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Player:      game.TargetPlayerReference(0),
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
										},
									},
									game.Instruction{
										Primitive: game.Draw{
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectReplacement("If a creature an opponent controls would die, exile it instead.", game.TriggerControllerAny, game.TriggerControllerOpponent, true, types.Creature),
			},
			OracleText: `
			If a creature an opponent controls would die, exile it instead.
			{2}, {T}, Sacrifice Stone of Erech: Exile target player's graveyard. Draw a card.
		`,
		},
	}
}
