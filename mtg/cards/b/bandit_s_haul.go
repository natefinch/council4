package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BanditSHaul is the card definition for Bandit's Haul.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Whenever you commit a crime, put a loot counter on this artifact. This ability triggers only once each turn. (Targeting opponents, anything they control, and/or cards in their graveyards is a crime.)
//	{T}: Add one mana of any color.
//	{2}, {T}, Remove two loot counters from this artifact: Draw a card.
var BanditSHaul = newBanditSHaul

func newBanditSHaul() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Bandit's Haul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Remove two loot counters from this artifact: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove two loot counters from this artifact",
							Amount:      2,
							CounterKind: counter.Loot,
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
							Event:  game.EventCrimeCommitted,
							Player: game.TriggerPlayerYou,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Loot,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you commit a crime, put a loot counter on this artifact. This ability triggers only once each turn. (Targeting opponents, anything they control, and/or cards in their graveyards is a crime.)
			{T}: Add one mana of any color.
			{2}, {T}, Remove two loot counters from this artifact: Draw a card.
		`,
		},
	}
}
