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
	whiteManaBatteryOracleID          = "cf58682f-c305-4803-b54b-37f0841788e9"
	whiteManaBatteryText              = "{2}, {T}: Put a charge counter on this artifact.\n{T}, Remove any number of charge counters from this artifact: Add {W}, then add an additional {W} for each charge counter removed this way."
	whiteManaBatteryCounterRemovalKey = "removed-charge-counters"
)

var WhiteManaBattery = cardset.Entry{
	OracleID: whiteManaBatteryOracleID,
	Name:     "White Mana Battery",
	New:      newWhiteManaBattery,
	Verification: "mtg/rules/colored_mana_batteries_test.go: TestWhiteManaBatteryStoresAndConvertsChargeCounters; " +
		"4ED printing b622e694-858d-482f-a67d-0e52d268708c; Scryfall ruling 2004-10-04; " +
		"CR 2026-09-25 106.3-4, 107.1c, 118.1-3, 122.1, 602.2a-b, 605.1a, 605.3a-b",
}

func newWhiteManaBattery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "White Mana Battery",
			ManaCost:   opt.Val(cost.Mana{cost.O(4)}),
			Types:      []types.Card{types.Artifact},
			OracleText: whiteManaBatteryText,
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
				Text: "{T}, Remove any number of charge counters from this artifact: Add {W}, then add an additional {W} for each charge counter removed this way.",
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{
						Kind:               cost.AdditionalRemoveCounter,
						CounterKind:        counter.Charge,
						ChooseCounterCount: true,
						SubjectKey:         whiteManaBatteryCounterRemovalKey,
					},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.AbilityContent{
					Modes: []game.Mode{{
						Sequence: []game.Instruction{
							{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.W}},
							{Primitive: game.AddMana{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind:    game.DynamicAmountPaidCostCounters,
									CostKey: whiteManaBatteryCounterRemovalKey,
								}),
								ManaColor: mana.W,
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
