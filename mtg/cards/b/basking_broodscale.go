package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BaskingBroodscale is the card definition for Basking Broodscale.
//
// Type: Creature — Eldrazi Lizard
// Cost: {1}{G}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	{1}{G}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)
//	Whenever one or more +1/+1 counters are put on this creature, you may create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
var BaskingBroodscale = newBaskingBroodscale()

func newBaskingBroodscale() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Basking Broodscale",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Lizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{G}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Adapt{
									Object: game.SourcePermanentReference(),
									Amount: game.Fixed(1),
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
							Event:            game.EventCountersAdded,
							Source:           game.TriggerSourceSelf,
							OneOrMore:        true,
							MatchCounterKind: true,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(baskingBroodscaleToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			{1}{G}: Adapt 1. (If this creature has no +1/+1 counters on it, put a +1/+1 counter on it.)
			Whenever one or more +1/+1 counters are put on this creature, you may create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
		`,
		},
	}
}

var baskingBroodscaleToken = newBaskingBroodscaleToken()

func newBaskingBroodscaleToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
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
		},
	}
}
