package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TomeOfLegends is the card definition for Tome of Legends.
//
// Type: Artifact — Book
// Cost: {2}
//
// Oracle text:
//
//	This artifact enters with a page counter on it.
//	Whenever your commander enters or attacks, put a page counter on this artifact.
//	{1}, {T}, Remove a page counter from this artifact: Draw a card.
var TomeOfLegends = newTomeOfLegends

func newTomeOfLegends() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tome of Legends",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Book},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, {T}, Remove a page counter from this artifact: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a page counter from this artifact",
							Amount:      1,
							CounterKind: counter.Page,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							UnionEvent:       game.EventAttackerDeclared,
							SubjectSelection: game.Selection{MatchCommander: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Page,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with a page counter on it.", game.CounterPlacement{Kind: counter.Page, Amount: 1}),
			},
			OracleText: `
			This artifact enters with a page counter on it.
			Whenever your commander enters or attacks, put a page counter on this artifact.
			{1}, {T}, Remove a page counter from this artifact: Draw a card.
		`,
		},
	}
}
