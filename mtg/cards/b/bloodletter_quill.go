package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BloodletterQuill is the card definition for Bloodletter Quill.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{2}, {T}, Put a blood counter on this artifact: Draw a card, then you lose 1 life for each blood counter on this artifact.
//	{U}{B}: Remove a blood counter from this artifact.
var BloodletterQuill = newBloodletterQuill()

func newBloodletterQuill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Bloodletter Quill",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Put a blood counter on this artifact: Draw a card, then you lose 1 life for each blood counter on this artifact.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalPutCounter,
							Text:        "Put a blood counter on this artifact",
							Amount:      1,
							CounterKind: counter.Blood,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  1,
										CounterKind: counter.Blood,
										Object:      game.SourcePermanentReference(),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{U}{B}: Remove a blood counter from this artifact.",
					ManaCost:       opt.Val(cost.Mana{cost.U, cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Blood,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}, {T}, Put a blood counter on this artifact: Draw a card, then you lose 1 life for each blood counter on this artifact.
			{U}{B}: Remove a blood counter from this artifact.
		`,
		},
	}
}
