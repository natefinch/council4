package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	solarionOracleID      = "4022e540-8fa0-4eee-bcf1-df5046289f2b"
	skyreachMantaOracleID = "58eab745-9afa-4db7-9d8a-f5eb0b9dca6c"
	pentadPrismOracleID   = "72e8d67a-41e6-4074-bcbd-b54ead995237"
	suntouchedMyrOracleID = "75d4f61a-8276-43ad-99e6-0d756a597e4b"
	etchedOracleOracleID  = "7ecfa47e-1165-46a6-884c-290b1c14d020"

	solarionOracleText = "Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)\n" +
		"{T}: Double the number of +1/+1 counters on this creature."
	skyreachMantaOracleText = "Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)\n" +
		"Flying"
	pentadPrismOracleText = "Sunburst (This artifact enters with a charge counter on it for each color of mana spent to cast it.)\n" +
		"Remove a charge counter from this artifact: Add one mana of any color."
	suntouchedMyrOracleText = "Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)"
	etchedOracleOracleText  = "Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)\n" +
		"{1}, Remove four +1/+1 counters from this creature: Target player draws three cards."
)

var (
	Solarion = cardset.Entry{
		OracleID: solarionOracleID,
		Name:     "Solarion",
		New:      newSolarion,
		Verification: "mtg/rules/sunburst_family_test.go: " +
			"TestSolarionSunburstCountersAndActivationDoubleThem; printing 9070dd3e-7600-493e-84b9-931ac92f44bc",
	}
	SkyreachManta = cardset.Entry{
		OracleID: skyreachMantaOracleID,
		Name:     "Skyreach Manta",
		New:      newSkyreachManta,
		Verification: "mtg/rules/sunburst_family_test.go: " +
			"TestSkyreachMantaSunburstAndFlying; printing 3df1e8a0-6a54-4be9-bd0c-ac70f0ae9aa6",
	}
	PentadPrism = cardset.Entry{
		OracleID: pentadPrismOracleID,
		Name:     "Pentad Prism",
		New:      newPentadPrism,
		Verification: "mtg/rules/sunburst_family_test.go: " +
			"TestPentadPrismSunburstAndChosenColorMana; printing e4beb84e-9eb9-48a2-932a-91fbbf218b5b",
	}
	SuntouchedMyr = cardset.Entry{
		OracleID: suntouchedMyrOracleID,
		Name:     "Suntouched Myr",
		New:      newSuntouchedMyr,
		Verification: "mtg/rules/sunburst_family_test.go: " +
			"TestSuntouchedMyrSunburstCounters; printing 4c5d4ae3-bc7e-4e60-a5e6-c3ae349fe6fc",
	}
	EtchedOracle = cardset.Entry{
		OracleID: etchedOracleOracleID,
		Name:     "Etched Oracle",
		New:      newEtchedOracle,
		Verification: "mtg/rules/sunburst_family_test.go: " +
			"TestEtchedOracleSunburstAndTargetedDrawActivation; printing 1b4e70c6-8f74-4b0b-9ad7-33fce05c401b",
	}
)

func sunburstReplacement(text string, kind counter.Kind) game.ReplacementAbility {
	return game.EntersWithCountersReplacement(text, game.CounterPlacement{
		Kind: kind,
		Dynamic: opt.Val(&game.DynamicAmount{
			Kind:       game.DynamicAmountColorsOfManaSpentToCast,
			Multiplier: 1,
		}),
	})
}

func newSolarion() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Solarion",
		ManaCost:   opt.Val(cost.Mana{cost.O(7)}),
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Construct},
		Power:      opt.Val(game.PT{Value: 0}),
		Toughness:  opt.Val(game.PT{Value: 0}),
		OracleText: solarionOracleText,
		ReplacementAbilities: []game.ReplacementAbility{
			sunburstReplacement("Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)", counter.PlusOnePlusOne),
		},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "{T}: Double the number of +1/+1 counters on this creature.",
			ZoneOfFunction:  zone.Battlefield,
			AdditionalCosts: cost.Tap,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddCounter{
					Object:      game.SourcePermanentReference(),
					CounterKind: counter.PlusOnePlusOne,
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:        game.DynamicAmountObjectCounters,
						Object:      game.SourcePermanentReference(),
						CounterKind: counter.PlusOnePlusOne,
						Multiplier:  1,
					}),
				},
			}}}.Ability(),
		}},
	}}
}

func newSkyreachManta() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Skyreach Manta",
		ManaCost:   opt.Val(cost.Mana{cost.O(5)}),
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Fish},
		Power:      opt.Val(game.PT{Value: 0}),
		Toughness:  opt.Val(game.PT{Value: 0}),
		OracleText: skyreachMantaOracleText,
		ReplacementAbilities: []game.ReplacementAbility{
			sunburstReplacement("Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)", counter.PlusOnePlusOne),
		},
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}}
}

func newPentadPrism() *game.CardDef {
	const manaChoice = game.ChoiceKey("pentad-prism-mana-color")
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Pentad Prism",
		ManaCost:   opt.Val(cost.Mana{cost.O(2)}),
		Types:      []types.Card{types.Artifact},
		OracleText: pentadPrismOracleText,
		ReplacementAbilities: []game.ReplacementAbility{
			sunburstReplacement("Sunburst (This artifact enters with a charge counter on it for each color of mana spent to cast it.)", counter.Charge),
		},
		ManaAbilities: []game.ManaAbility{{
			Text: "{Remove a charge counter from this artifact}: Add one mana of any color.",
			AdditionalCosts: []cost.Additional{{
				Kind:        cost.AdditionalRemoveCounter,
				Text:        "Remove a charge counter from this artifact",
				Amount:      1,
				CounterKind: counter.Charge,
			}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.Choose{
					Choice: game.ResolutionChoice{
						Kind:   game.ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
					},
					PublishChoice: manaChoice,
				}},
				{Primitive: game.AddMana{Amount: game.Fixed(1), ChoiceFrom: manaChoice}},
			}}.Ability(),
		}},
	}}
}

func newSuntouchedMyr() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Suntouched Myr",
		ManaCost:   opt.Val(cost.Mana{cost.O(3)}),
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Myr},
		Power:      opt.Val(game.PT{Value: 0}),
		Toughness:  opt.Val(game.PT{Value: 0}),
		OracleText: suntouchedMyrOracleText,
		ReplacementAbilities: []game.ReplacementAbility{
			sunburstReplacement("Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)", counter.PlusOnePlusOne),
		},
	}}
}

func newEtchedOracle() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Etched Oracle",
		ManaCost:   opt.Val(cost.Mana{cost.O(4)}),
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Wizard},
		Power:      opt.Val(game.PT{Value: 0}),
		Toughness:  opt.Val(game.PT{Value: 0}),
		OracleText: etchedOracleOracleText,
		ReplacementAbilities: []game.ReplacementAbility{
			sunburstReplacement("Sunburst (This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.)", counter.PlusOnePlusOne),
		},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:           "{1}, Remove four +1/+1 counters from this creature: Target player draws three cards.",
			ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
			ZoneOfFunction: zone.Battlefield,
			AdditionalCosts: []cost.Additional{{
				Kind:        cost.AdditionalRemoveCounter,
				Text:        "Remove four +1/+1 counters from this creature",
				Amount:      4,
				CounterKind: counter.PlusOnePlusOne,
			}},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target player",
					Allow:      game.TargetAllowPlayer,
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(3), Player: game.TargetPlayerReference(0)},
				}},
			}.Ability(),
		}},
	}}
}
