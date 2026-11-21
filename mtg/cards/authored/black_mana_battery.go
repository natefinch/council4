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
	blackManaBatteryOracleID          = "dc3b2bab-755d-4b2f-97c0-70a74d721a79"
	blackManaBatteryText              = "{2}, {T}: Put a charge counter on this artifact.\n{T}, Remove any number of charge counters from this artifact: Add {B}, then add an additional {B} for each charge counter removed this way."
	blackManaBatteryCounterRemovalKey = "removed-charge-counters"
)

var BlackManaBattery = cardset.Entry{
	OracleID: blackManaBatteryOracleID,
	Name:     "Black Mana Battery",
	New:      newBlackManaBattery,
	Verification: "mtg/rules/colored_mana_batteries_test.go: TestBlackManaBatteryStoresAndConvertsChargeCounters; " +
		"4ED printing a81f7a0f-183f-438b-b252-738d8d30c245; Scryfall ruling 2004-10-04; " +
		"CR 2026-09-25 106.3-4, 107.1c, 118.1-3, 122.1, 602.2a-b, 605.1a, 605.3a-b",
}

func newBlackManaBattery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Black Mana Battery",
			ManaCost:   opt.Val(cost.Mana{cost.O(4)}),
			Types:      []types.Card{types.Artifact},
			OracleText: blackManaBatteryText,
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
				Text: "{T}, Remove any number of charge counters from this artifact: Add {B}, then add an additional {B} for each charge counter removed this way.",
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{
						Kind:               cost.AdditionalRemoveCounter,
						CounterKind:        counter.Charge,
						ChooseCounterCount: true,
						SubjectKey:         blackManaBatteryCounterRemovalKey,
					},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.AbilityContent{
					Modes: []game.Mode{{
						Sequence: []game.Instruction{
							{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.B}},
							{Primitive: game.AddMana{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind:    game.DynamicAmountPaidCostCounters,
									CostKey: blackManaBatteryCounterRemovalKey,
								}),
								ManaColor: mana.B,
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
