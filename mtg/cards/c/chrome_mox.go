package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChromeMox is the card definition for Chrome Mox.
//
// Type: Artifact
// Cost: {0}
//
// Oracle text:
//
//	Imprint — When this artifact enters, you may exile a nonartifact, nonland card from your hand.
//	{T}: Add one mana of any of the exiled card's colors.
var ChromeMox = newChromeMox

func newChromeMox() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Chrome Mox",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapLinkedExileColorManaAbility("imprint"),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Imprint — When this artifact enters, you may exile a nonartifact, nonland card from your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Hand,
											Filter:     game.Selection{ExcludedTypes: []types.Card{types.Artifact, types.Land}},
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
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Imprint — When this artifact enters, you may exile a nonartifact, nonland card from your hand.
			{T}: Add one mana of any of the exiled card's colors.
		`,
		},
	}
}
