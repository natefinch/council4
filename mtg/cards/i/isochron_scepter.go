package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IsochronScepter is the card definition for Isochron Scepter.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Imprint — When this artifact enters, you may exile an instant card with mana value 2 or less from your hand.
//	{2}, {T}: You may copy the exiled card. If you do, you may cast the copy without paying its mana cost.
var IsochronScepter = newIsochronScepter

func newIsochronScepter() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Isochron Scepter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}, {T}: You may copy the exiled card. If you do, you may cast the copy without paying its mana cost.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CopyCard{
									Player: game.ControllerReference(),
									LinkID: "imprint",
								},
								Optional:      true,
								PublishResult: game.ResultKey("imprint-copy-made"),
							},
							{
								Primitive: game.PlayLinkedExiledCard{
									Player:                game.ControllerReference(),
									LinkID:                "imprint",
									Copy:                  true,
									WithoutPayingManaCost: true,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "imprint-copy-made",
									Succeeded: game.TriTrue,
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Instant}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Exile,
									},
									Riders: game.ChooseRiders{
										PublishLinked:       game.LinkedKey("imprint"),
										PublishObjectScoped: true,
									},
									Prompt: "Choose a card to exile",
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Imprint — When this artifact enters, you may exile an instant card with mana value 2 or less from your hand.
			{2}, {T}: You may copy the exiled card. If you do, you may cast the copy without paying its mana cost.
		`,
		},
	}
}
