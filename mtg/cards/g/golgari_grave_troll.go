package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GolgariGraveTroll is the card definition for Golgari Grave-Troll.
//
// Type: Creature — Troll Skeleton
// Cost: {4}{G}
//
// Oracle text:
//
//	This creature enters with a +1/+1 counter on it for each creature card in your graveyard.
//	{1}, Remove a +1/+1 counter from this creature: Regenerate this creature.
//	Dredge 6 (If you would draw a card, you may mill six cards instead. If you do, return this card from your graveyard to your hand.)
var GolgariGraveTroll = newGolgariGraveTroll

func newGolgariGraveTroll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Golgari Grave-Troll",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Troll, types.Skeleton},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.DredgeStaticAbility(6),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, Remove a +1/+1 counter from this creature: Regenerate this creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a +1/+1 counter from this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Regenerate{
											Object: game.SourcePermanentReference(),
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
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each creature card in your graveyard.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountCountCardsInZone,
					Multiplier: 1,
					Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
					CardZone:   zone.Graveyard,
					Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
				})}),
			},
			OracleText: `
			This creature enters with a +1/+1 counter on it for each creature card in your graveyard.
			{1}, Remove a +1/+1 counter from this creature: Regenerate this creature.
			Dredge 6 (If you would draw a card, you may mill six cards instead. If you do, return this card from your graveyard to your hand.)
		`,
		},
	}
}
