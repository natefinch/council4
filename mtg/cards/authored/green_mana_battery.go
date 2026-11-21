package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	greenManaBatteryOracleID          = "b953d19f-6ba7-4aee-b353-c60dc9572610"
	greenManaBatteryText              = "{2}, {T}: Put a charge counter on this artifact.\n{T}, Remove any number of charge counters from this artifact: Add {G}, then add an additional {G} for each charge counter removed this way."
	greenManaBatteryCounterRemovalKey = "removed-charge-counters"
)

var GreenManaBattery = cardset.Entry{
	OracleID: greenManaBatteryOracleID,
	Name:     "Green Mana Battery",
	New:      newGreenManaBattery,
	Verification: "mtg/rules/colored_mana_batteries_test.go: TestGreenManaBatteryStoresAndConvertsChargeCounters; " +
		"4ED printing d0a6e224-72df-4f1f-93d1-5114779aed2c; Scryfall ruling 2004-10-04; " +
		"CR 2026-09-25 106.3-4, 107.1c, 118.1-3, 122.1, 602.2a-b, 605.1a, 605.3a-b",
}

func newGreenManaBattery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Green Mana Battery",
			ManaCost:   opt.Val(cost.Mana{cost.O(4)}),
			Types:      []types.Card{types.Artifact},
			OracleText: greenManaBatteryText,
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:            "{2}, {T}: Put a charge counter on this artifact.",
				ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
				AdditionalCosts: cost.Tap,
				ZoneOfFunction:  zone.Battlefield,
				Content: game.Mode{
					Sequence: []game.Instruction{{
						Primitive: game.AddCounter{
							Object:      game.SourcePermanentReference(),
							Amount:      game.Fixed(1),
							CounterKind: counter.Charge,
						},
					}},
				}.Ability(),
			}},
			ManaAbilities: []game.ManaAbility{{
				Text: "{T}, Remove any number of charge counters from this artifact: Add {G}, then add an additional {G} for each charge counter removed this way.",
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{
						Kind:               cost.AdditionalRemoveCounter,
						CounterKind:        counter.Charge,
						ChooseCounterCount: true,
						SubjectKey:         greenManaBatteryCounterRemovalKey,
					},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.AbilityContent{
					Modes: []game.Mode{{
						Sequence: []game.Instruction{
							{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
							{Primitive: game.AddMana{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind:    game.DynamicAmountPaidCostCounters,
									CostKey: greenManaBatteryCounterRemovalKey,
								}),
								ManaColor: mana.G,
							}},
						},
					}},
					MinModes: 1,
					MaxModes: 1,
				},
			}},
		},
	}
}
