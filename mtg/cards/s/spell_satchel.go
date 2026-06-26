package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpellSatchel is the card definition for Spell Satchel.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Magecraft — Whenever you cast or copy an instant or sorcery spell, put a book counter on this artifact.
//	{T}, Remove a book counter from this artifact: Add {C}.
//	{3}, {T}, Remove three book counters from this artifact: Draw a card.
var SpellSatchel = newSpellSatchel()

func newSpellSatchel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Spell Satchel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}, {T}, Remove three book counters from this artifact: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove three book counters from this artifact",
							Amount:      3,
							CounterKind: counter.Book,
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
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a book counter from this artifact",
							Amount:      1,
							CounterKind: counter.Book,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
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
							Event:          game.EventSpellCast,
							Controller:     game.TriggerControllerYou,
							MatchSpellCopy: true,
							CardSelection:  game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Book,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Magecraft — Whenever you cast or copy an instant or sorcery spell, put a book counter on this artifact.
			{T}, Remove a book counter from this artifact: Add {C}.
			{3}, {T}, Remove three book counters from this artifact: Draw a card.
		`,
		},
	}
}
