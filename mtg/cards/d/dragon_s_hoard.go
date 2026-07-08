package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DragonSHoard is the card definition for Dragon's Hoard.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Whenever a Dragon you control enters, put a gold counter on this artifact.
//	{T}, Remove a gold counter from this artifact: Draw a card.
//	{T}: Add one mana of any color.
var DragonSHoard = newDragonSHoard

func newDragonSHoard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dragon's Hoard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove a gold counter from this artifact: Draw a card.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a gold counter from this artifact",
							Amount:      1,
							CounterKind: counter.Gold,
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
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Dragon")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Gold,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a Dragon you control enters, put a gold counter on this artifact.
			{T}, Remove a gold counter from this artifact: Draw a card.
			{T}: Add one mana of any color.
		`,
		},
	}
}
