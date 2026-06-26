package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RingOfThreeWishes is the card definition for Ring of Three Wishes.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	This artifact enters with three wish counters on it.
//	{5}, {T}, Remove a wish counter from this artifact: Search your library for a card, put that card into your hand, then shuffle.
var RingOfThreeWishes = newRingOfThreeWishes()

func newRingOfThreeWishes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ring of Three Wishes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{5}, {T}, Remove a wish counter from this artifact: Search your library for a card, put that card into your hand, then shuffle.",
					ManaCost: opt.Val(cost.Mana{cost.O(5)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a wish counter from this artifact",
							Amount:      1,
							CounterKind: counter.Wish,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:       zone.Library,
										Destination:      zone.Hand,
										FailToFindPolicy: game.SearchMustFindIfAvailable,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with three wish counters on it.", game.CounterPlacement{Kind: counter.Wish, Amount: 3}),
			},
			OracleText: `
			This artifact enters with three wish counters on it.
			{5}, {T}, Remove a wish counter from this artifact: Search your library for a card, put that card into your hand, then shuffle.
		`,
		},
	}
}
