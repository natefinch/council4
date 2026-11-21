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
	blueManaBatteryOracleID          = "79e91e8a-cee7-411b-b0a3-d3383c0d4a43"
	blueManaBatteryText              = "{2}, {T}: Put a charge counter on this artifact.\n{T}, Remove any number of charge counters from this artifact: Add {U}, then add an additional {U} for each charge counter removed this way."
	blueManaBatteryCounterRemovalKey = "removed-charge-counters"
)

var BlueManaBattery = cardset.Entry{
	OracleID: blueManaBatteryOracleID,
	Name:     "Blue Mana Battery",
	New:      newBlueManaBattery,
	Verification: "mtg/rules/blue_mana_battery_test.go: TestBlueManaBatteryAddsChargeCounter, " +
		"TestBlueManaBatteryProducesBlueManaFromChosenChargeCounters; printing cd2cb84e-c079-486e-87ad-d188fe38bc76; " +
		"Scryfall ruling 2004-10-04; CR 2026-09-25 106.3-4, 107.1c, 118.1-3, 602.2a-b, 605.1a, 605.3a-b",
}

func newBlueManaBattery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Blue Mana Battery",
			ManaCost:   opt.Val(cost.Mana{cost.O(4)}),
			Types:      []types.Card{types.Artifact},
			OracleText: blueManaBatteryText,
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
				Text: "{T}, Remove any number of charge counters from this artifact: Add {U}, then add an additional {U} for each charge counter removed this way.",
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{
						Kind:               cost.AdditionalRemoveCounter,
						CounterKind:        counter.Charge,
						ChooseCounterCount: true,
						SubjectKey:         blueManaBatteryCounterRemovalKey,
					},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.AbilityContent{
					Modes: []game.Mode{{
						Sequence: []game.Instruction{
							{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.U}},
							{Primitive: game.AddMana{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind:    game.DynamicAmountPaidCostCounters,
									CostKey: blueManaBatteryCounterRemovalKey,
								}),
								ManaColor: mana.U,
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
