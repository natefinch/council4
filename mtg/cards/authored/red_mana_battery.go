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
	redManaBatteryOracleID          = "8ceb295c-a611-4ac7-940b-557bf60f8a8b"
	redManaBatteryText              = "{2}, {T}: Put a charge counter on this artifact.\n{T}, Remove any number of charge counters from this artifact: Add {R}, then add an additional {R} for each charge counter removed this way."
	redManaBatteryCounterRemovalKey = "removed-charge-counters"
)

var RedManaBattery = cardset.Entry{
	OracleID: redManaBatteryOracleID,
	Name:     "Red Mana Battery",
	New:      newRedManaBattery,
	Verification: "mtg/rules/colored_mana_batteries_test.go: TestRedManaBatteryStoresAndConvertsChargeCounters; " +
		"4ED printing e4e507bc-441d-4f44-85d4-cf93ca199d2e; Scryfall ruling 2004-10-04; " +
		"CR 2026-09-25 106.3-4, 107.1c, 118.1-3, 122.1, 602.2a-b, 605.1a, 605.3a-b",
}

func newRedManaBattery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Red Mana Battery",
			ManaCost:   opt.Val(cost.Mana{cost.O(4)}),
			Types:      []types.Card{types.Artifact},
			OracleText: redManaBatteryText,
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
				Text: "{T}, Remove any number of charge counters from this artifact: Add {R}, then add an additional {R} for each charge counter removed this way.",
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{
						Kind:               cost.AdditionalRemoveCounter,
						CounterKind:        counter.Charge,
						ChooseCounterCount: true,
						SubjectKey:         redManaBatteryCounterRemovalKey,
					},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.AbilityContent{
					Modes: []game.Mode{{
						Sequence: []game.Instruction{
							{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.R}},
							{Primitive: game.AddMana{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind:    game.DynamicAmountPaidCostCounters,
									CostKey: redManaBatteryCounterRemovalKey,
								}),
								ManaColor: mana.R,
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
