package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BalmOfRestoration is the card definition for Balm of Restoration.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{1}, {T}, Sacrifice this artifact: Choose one —
//	• You gain 2 life.
//	• Prevent the next 2 damage that would be dealt to any target this turn.
var BalmOfRestoration = newBalmOfRestoration()

func newBalmOfRestoration() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Balm of Restoration",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: `{1}, {T}, Sacrifice this artifact: Choose one —
		• You gain 2 life.
		• Prevent the next 2 damage that would be dealt to any target this turn.`,
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "You gain 2 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Prevent the next 2 damage that would be dealt to any target this turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "any target",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.PreventDamage{
											AnyTarget: game.AnyTargetDamageRecipient(0),
											Amount:    game.Fixed(2),
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
			{1}, {T}, Sacrifice this artifact: Choose one —
			• You gain 2 life.
			• Prevent the next 2 damage that would be dealt to any target this turn.
		`,
		},
	}
}
